package provider

import (
	"context"
	"errors"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type contentfulRequestMethodContextKey struct{}

type contentfulRequestNoRetryContextKey struct{}

type contentfulRequestRetryStateContextKey struct{}

type contentfulRequestRetryState struct {
	nextBackoffAttempt int
	backoffAttempt     int
	backoff            time.Duration
	backoffReady       bool
}

type contentfulRetryCoordinator struct {
	client *retryablehttp.Client
}

type contentfulRequestContextRoundTripper struct {
	next http.RoundTripper
}

type contentfulRequestCancelReadCloser struct {
	io.ReadCloser

	cancel context.CancelFunc
}

func (r contentfulRequestCancelReadCloser) Read(bytes []byte) (int, error) {
	read, err := r.ReadCloser.Read(bytes)
	if err != nil {
		r.cancel()
	}

	return read, err //nolint:wrapcheck // preserve response body read errors.
}

func (r contentfulRequestCancelReadCloser) Close() error {
	defer r.cancel()

	return r.ReadCloser.Close() //nolint:wrapcheck // preserve response body close errors.
}

func (t contentfulRequestContextRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	ctx := request.Context()

	err := ctx.Err()
	if err != nil {
		return nil, err //nolint:wrapcheck // preserve context cancellation identity for net/http.
	}

	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, defaultResourceOperationTimeout)
	}

	ctx = context.WithValue(ctx, contentfulRequestMethodContextKey{}, request.Method)
	ctx = context.WithValue(ctx, contentfulRequestRetryStateContextKey{}, &contentfulRequestRetryState{})

	// Preserve transport errors for net/http and the retry classifier.
	response, err := t.next.RoundTrip(request.WithContext(ctx))
	if cancel == nil {
		return response, err //nolint:wrapcheck
	}

	if err != nil || response == nil || response.Body == nil {
		cancel()

		return response, err //nolint:wrapcheck
	}

	response.Body = contentfulRequestCancelReadCloser{ReadCloser: response.Body, cancel: cancel}

	return response, nil
}

func newContentfulHTTPClient(httpClient *http.Client) *http.Client {
	retryableClient := retryablehttp.NewClient()
	retryableClient.Logger = nil
	retryCoordinator := contentfulRetryCoordinator{
		client: retryableClient,
	}
	retryableClient.Backoff = retryCoordinator.backoff
	retryableClient.CheckRetry = retryCoordinator.checkRetry
	retryableClient.ErrorHandler = retryablehttp.PassthroughErrorHandler
	retryableClient.RetryMax = math.MaxInt

	// Operation contexts, including the provider default added by the outer
	// transport when needed, are the retry budget. RetryMax remains only because
	// retryablehttp requires an integer limit; it is not an effective bound.
	if httpClient != nil {
		clientCopy := *httpClient
		retryableClient.HTTPClient = &clientCopy
	}

	retryableClient.HTTPClient.CheckRedirect = contentfulCheckRedirect(retryableClient.HTTPClient.CheckRedirect)

	client := retryableClient.StandardClient()
	client.CheckRedirect = contentfulCheckRedirect(client.CheckRedirect)
	client.Transport = contentfulRequestContextRoundTripper{next: client.Transport}

	return client
}

func (c contentfulRetryCoordinator) checkRetry(
	ctx context.Context,
	response *http.Response,
	err error,
) (bool, error) {
	shouldRetry, err := contentfulRetryPolicy(ctx, response, err)
	if err != nil || !shouldRetry {
		return shouldRetry, err
	}

	retryState, ok := ctx.Value(contentfulRequestRetryStateContextKey{}).(*contentfulRequestRetryState)
	if !ok {
		return true, nil
	}

	attempt := retryState.nextBackoffAttempt
	retryState.nextBackoffAttempt++

	if response == nil {
		return true, nil
	}

	backoff := util.ContentfulRateLimitLinearJitterBackoff(
		c.client.RetryWaitMin,
		c.client.RetryWaitMax,
		attempt,
		response,
	)

	err = ctx.Err()
	if err != nil {
		return false, err //nolint:wrapcheck // preserve authoritative context cancellation identity.
	}

	method, _ := ctx.Value(contentfulRequestMethodContextKey{}).(string)
	logFields := map[string]any{
		"decision":      "retry",
		"method":        method,
		"status_code":   response.StatusCode,
		"retry_ordinal": attempt + 1,
		"wait_ms":       backoff.Milliseconds(),
	}

	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return false, context.DeadlineExceeded
		}

		logFields["deadline_remaining_ms"] = remaining.Milliseconds()

		if backoff >= remaining {
			logFields["decision"] = "decline_deadline"
			tflog.Debug(ctx, "contentful.response_retry", logFields)

			return false, nil
		}
	}

	tflog.Debug(ctx, "contentful.response_retry", logFields)

	retryState.backoffAttempt = attempt
	retryState.backoff = backoff
	retryState.backoffReady = true

	return true, nil
}

func (c contentfulRetryCoordinator) backoff(
	minDelay time.Duration,
	maxDelay time.Duration,
	attempt int,
	response *http.Response,
) time.Duration {
	if response != nil && response.Request != nil {
		retryState, ok := response.Request.Context().Value(contentfulRequestRetryStateContextKey{}).(*contentfulRequestRetryState)
		if ok && retryState.backoffReady && retryState.backoffAttempt == attempt {
			retryState.backoffReady = false

			return retryState.backoff
		}
	}

	return util.ContentfulRateLimitLinearJitterBackoff(minDelay, maxDelay, attempt, response)
}

func contentfulRetryPolicy(ctx context.Context, response *http.Response, err error) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err() //nolint:wrapcheck // retryablehttp recognizes context cancellation by identity.
	}

	if noRetry, _ := ctx.Value(contentfulRequestNoRetryContextKey{}).(bool); noRetry {
		// Selected mutation lifecycles use the request context to prohibit every
		// transparent replay, including explicit 429 responses. The signal is
		// intentionally narrower than the HTTP method because unrelated CMA
		// mutations retain the provider's default retry policy.
		return false, nil
	}

	if err == nil && response != nil && response.StatusCode == http.StatusTooManyRequests {
		// Follow Contentful's documented and first-party 429 retry practice for
		// every method; this does not prove that a mutation was uncommitted.
		return true, nil
	}

	method, _ := ctx.Value(contentfulRequestMethodContextKey{}).(string)

	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return retryablehttp.DefaultRetryPolicy(ctx, response, err) //nolint:wrapcheck
	default:
		// A transport error or ordinary server response cannot establish
		// whether Contentful committed a mutation. Transparent replay could
		// repeat a committed write under the same optimistic-lock version.
		return false, nil
	}
}

func withContentfulRequestNoRetry(ctx context.Context) context.Context {
	return context.WithValue(ctx, contentfulRequestNoRetryContextKey{}, true)
}

const contentfulDefaultRedirectLimit = 10

var errContentfulRedirectLimit = errors.New("stopped after 10 redirects")

// Both nested HTTP clients must honor the mutation replay boundary. Otherwise
// the outer client can follow a redirect that the inner client declined.
func contentfulCheckRedirect(policy func(*http.Request, []*http.Request) error) func(*http.Request, []*http.Request) error {
	return func(request *http.Request, via []*http.Request) error {
		if noRetry, _ := request.Context().Value(contentfulRequestNoRetryContextKey{}).(bool); noRetry {
			return http.ErrUseLastResponse
		}

		if policy != nil {
			return policy(request, via)
		}

		// Match net/http's default when the supplied client has no custom policy.
		if len(via) >= contentfulDefaultRedirectLimit {
			return errContentfulRedirectLimit
		}

		return nil
	}
}
