//nolint:testpackage // Package-local lifecycle tests share this request counter.
package provider

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/require"
)

func mutationRequestCountingClient(t *testing.T) (*cm.Client, *atomic.Int64) {
	t.Helper()

	requestCount := &atomic.Int64{}
	testServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, _ *http.Request) {
		requestCount.Add(1)
		http.Error(responseWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}))
	t.Cleanup(testServer.Close)

	client, err := cm.NewClient(
		testServer.URL,
		cm.NewAccessTokenSecuritySource("access-token"),
		cm.WithClient(testServer.Client()),
	)
	require.NoError(t, err)

	return client, requestCount
}
