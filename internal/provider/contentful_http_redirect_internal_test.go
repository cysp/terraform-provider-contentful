package provider

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentfulNoRetryRejectsRedirectsInBothClients(t *testing.T) {
	t.Parallel()

	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		for _, status := range []int{301, 302, 303, 307, 308} {
			t.Run(method+"/"+http.StatusText(status), func(t *testing.T) {
				t.Parallel()

				var destinationCount atomic.Int64

				destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					destinationCount.Add(1)
					w.WriteHeader(http.StatusOK)
				}))
				t.Cleanup(destination.Close)

				var sourceCount atomic.Int64

				source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					sourceCount.Add(1)
					assert.Equal(t, method, req.Method)
					http.Redirect(w, req, destination.URL, status)
				}))
				t.Cleanup(source.Close)
				baseClient := source.Client()
				client := newContentfulHTTPClient(baseClient)
				req, err := http.NewRequestWithContext(withContentfulRequestNoRetry(t.Context()), method, source.URL, strings.NewReader(`{"value":"DO_NOT_FORWARD_SECRET"}`))
				require.NoError(t, err)
				resp, err := client.Do(req)
				require.NoError(t, err)
				require.NoError(t, resp.Body.Close())
				assert.Equal(t, status, resp.StatusCode)
				assert.EqualValues(t, 1, sourceCount.Load())
				assert.Zero(t, destinationCount.Load(), "neither inner nor outer client may follow")
				assert.Nil(t, baseClient.CheckRedirect, "caller-owned client must remain unchanged")
			})
		}
	}
}

func TestContentfulRedirectPoliciesForOrdinaryRequests(t *testing.T) {
	t.Parallel()

	var count atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		count.Add(1)

		if req.URL.Path == "/start" {
			http.Redirect(w, req, "/final", http.StatusTemporaryRedirect)

			return
		}

		if req.URL.Path == "/loop" {
			http.Redirect(w, req, "/loop", http.StatusTemporaryRedirect)

			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	client := newContentfulHTTPClient(server.Client())
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+"/start", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.EqualValues(t, 2, count.Load())

	count.Store(0)

	req, err = http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+"/loop", nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	if resp != nil {
		require.NoError(t, resp.Body.Close())
	}

	require.ErrorIs(t, err, errContentfulRedirectLimit)
	assert.EqualValues(t, 10, count.Load())

	var customCount atomic.Int64

	base := server.Client()
	base.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		customCount.Add(1)

		return http.ErrUseLastResponse
	}
	client = newContentfulHTTPClient(base)
	// Call the provider's transport to observe the inner custom policy without
	// the outer client's independently existing default behavior.
	req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/start", nil)
	require.NoError(t, err)
	resp, err = client.Transport.RoundTrip(req)
	require.NoError(t, err)
	_, err = io.Copy(io.Discard, resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	assert.EqualValues(t, 1, customCount.Load())
	require.ErrorIs(t, base.CheckRedirect(nil, nil), http.ErrUseLastResponse)
}
