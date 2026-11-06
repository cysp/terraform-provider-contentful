package provider

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpacesDataSourceOrganizationScopeContradiction(t *testing.T) {
	t.Parallel()
	body := discoveryFixture(t, "space")
	response := discoveryReadTest(t.Context(), t, NewSpacesDataSource, map[string]any{"organization_id": "another-org"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return discoveryHTTPResponse(request, 200, discoveryPage(0, 100, 1, body)), nil
	}))
	assert.True(t, response.Diagnostics.HasError())
	assert.True(t, response.State.Raw.IsNull())
}
