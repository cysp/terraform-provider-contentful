# Contentful HTTP retry policy

Read this contract when changing retries, deadlines, or mutation recovery.
It defines the provider HTTP layer's automatic retry and deadline boundaries
for Contentful Management API (CMA) requests. The implementation is in
[`contentful_http_client.go`](../../internal/provider/contentful_http_client.go).

| Request | Explicit 429 | Retryable transport failure or server response |
| --- | --- | --- |
| GET, HEAD, OPTIONS | Retry within the deadline | Retry within the deadline |
| POST, PUT, PATCH, DELETE by default | Retry within the deadline | Return the result without replay |
| Entry Create, specified-ID Create, Update, Publish; Content Type Create, Update, Activate | Return the first result without replay | Return the first result without replay |

The response deadline and evidence limits below are part of this policy.

## Deadline budget

Every CMA request has a finite context deadline. Resource and data-source
operations establish their configured Terraform timeout, normally defaulting
to two minutes, before calling the CMA client. Those operation deadlines remain
authoritative. If another call path, such as a list resource or internal CMA
operation, reaches the provider HTTP layer without a deadline, that layer adds
the same two-minute provider default to the request.

The context deadline or cancellation is the effective retry budget. The HTTP
layer does not expose or enforce a practitioner-configurable retry count.
Backoff waits select on the request context, so a long Contentful reset value
cannot extend an operation past its deadline.

Before entering a retry wait for a retryable HTTP response, the provider
calculates the normal backoff. If the context is still active but that
response-retry wait cannot complete before its deadline, the provider declines
the retry immediately and returns the final HTTP response without draining it.
This differs from a context that has already expired or been cancelled, or that
expires or is cancelled during a retry wait: in those cases the context error
remains authoritative.

## Retry classification

By default, the provider retries:

- explicit HTTP 429 responses for every method, following Contentful's
  documented rate-limit handling and first-party client practice; and
- eligible transport failures and server responses for safe GET, HEAD, and
  OPTIONS requests.

For safe methods, eligibility follows the pinned
[`retryablehttp.DefaultRetryPolicy`](https://github.com/hashicorp/go-retryablehttp/blob/v0.7.8/client.go#L472-L543).
It retries most transport errors and server errors other than 501. It excludes
recognized certificate-verification failures, invalid schemes or headers, and
exhausted redirects. Context cancellation and deadline expiry always stop
retrying. The [dependency pin](../../go.mod) determines this classification.

The provider does not transparently replay POST, PUT, PATCH, or DELETE after a
transport failure or an ordinary 5xx response. Those outcomes do not establish
whether Contentful committed the mutation, so replay could repeat an already
applied write.

### Rate-limit evidence

| Evidence | Establishes | Does not establish |
| --- | --- | --- |
| Contentful's [CMA rate-limit documentation](https://www.contentful.com/developers/docs/references/content-management-api/overview/#api-rate-limits) | 429 represents rate limiting and the reset interval tells clients when to retry | Whether a mutation returning 429 committed |
| Contentful's [first-party management SDK](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/README.md#L387-L389) | The SDK retries 429 and 500 responses by default | A server commitment guarantee |

The Entry Create, specified-ID Create, Update, and Publish calls and the Content
Type Create, Update, and Activate calls opt out of transparent retry for the
complete request. For those exact lifecycle mutations, explicit 429 responses,
transport failures, and 5xx responses are returned after one request. The
private request-context signal is checked before the general all-method 429
branch and survives generated-client request construction. GET and unrelated
CMA operations retain the default policy.

## Backoff and final errors

For a valid `X-Contentful-RateLimit-Reset` response, the reset is the earliest
retry time. The provider waits for the reset plus 100ms and full jitter from a
contention window that starts at 500ms, doubles for each retryablehttp backoff
attempt, and caps at four seconds. The reset value itself is never multiplied.
Missing, invalid, or unrepresentable reset values retain retryablehttp's
`Retry-After` and linear-jitter fallback behavior.

The backoff implementation is
[`ContentfulRateLimitLinearJitterBackoff`](../../internal/provider/util/retry_backoff.go).

If retryablehttp reaches its terminal error-handler path, the final HTTP
response or underlying transport error is passed through rather than replaced
with a generic retry-exhaustion error. Cancellation or deadline expiry during a
backoff wait instead returns the context error and no prior 429 response.

## Safety boundary

For POST, PUT, PATCH, and DELETE generally, this policy prevents transparent
replay after transport failures and ordinary 5xx responses while retaining the
documented default 429 behavior. The narrower Entry-publication and Content
Type-activation lifecycle boundary also disables 429 replay because an
ambiguous mutation cannot safely create or replace exact-version authority.

This does not provide at-most-once semantics across separate Terraform
operations or applies. In particular, repeating a create whose first result was
ambiguous can create another remote object when Contentful generates the
identifier. A validated draft response may instead authorize later recovery of
only its exact returned version; a failed or ambiguous draft response grants no
such authority.
