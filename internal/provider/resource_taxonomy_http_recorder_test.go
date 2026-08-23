package provider_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaxonomyRequestBodyRecorderExcludesRateLimitedAttempts(t *testing.T) {
	t.Parallel()

	responses := []int{http.StatusTooManyRequests, http.StatusConflict, http.StatusNoContent}
	next := http.HandlerFunc(func(responseWriter http.ResponseWriter, _ *http.Request) {
		responseWriter.WriteHeader(responses[0])
		responses = responses[1:]
	})
	recorder := &taxonomyRequestBodyRecorder{next: next}

	requests := []struct {
		status  int
		version string
		body    string
	}{
		{status: http.StatusTooManyRequests, version: "1", body: `[{"attempt":"rate-limited"}]`},
		{status: http.StatusConflict, version: "2", body: `[{"attempt":"conflict"}]`},
		{status: http.StatusNoContent, version: "3", body: `[{"attempt":"processed"}]`},
	}

	for _, expected := range requests {
		request := httptest.NewRequestWithContext(t.Context(), http.MethodPatch, "/organizations/organization-id/taxonomy/concepts/furniture", strings.NewReader(expected.body))
		request.Header.Set("X-Contentful-Version", expected.version)

		response := httptest.NewRecorder()
		recorder.ServeHTTP(response, request)
		require.Equal(t, expected.status, response.Code)
	}

	processed := recorder.matchingRequests(http.MethodPatch, "/taxonomy/concepts/furniture")
	require.Len(t, processed, 2)
	assert.Equal(t, "2", processed[0].version)
	assert.JSONEq(t, `[{"attempt":"conflict"}]`, string(processed[0].body))
	assert.Equal(t, "3", processed[1].version)
	assert.JSONEq(t, `[{"attempt":"processed"}]`, string(processed[1].body))
}
