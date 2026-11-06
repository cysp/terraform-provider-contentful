package provider

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocaleDataSourceFallbackValues(t *testing.T) {
	t.Parallel()

	body := discoveryFixture(t, "locale")
	for _, testcase := range []struct {
		name     string
		change   func(map[string]any)
		expected string
	}{
		{"absent", func(document map[string]any) { delete(document, "fallbackCode") }, testJSON(nil)},
		{"null", func(document map[string]any) { document["fallbackCode"] = nil }, testJSON(nil)},
		{"empty", func(document map[string]any) { document["fallbackCode"] = "" }, testJSON("")},
		{"code", func(document map[string]any) { document["fallbackCode"] = "fr-FR" }, testJSON("fr-FR")},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			response := discoveryReadTest(t.Context(), t, NewLocaleDataSource, map[string]any{"space_id": "space", "environment_id": "master", "locale_id": "item-b"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
				return discoveryHTTPResponse(request, 200, mutateTestJSON(body, testcase.change)), nil
			}))
			require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
			discoveryStateAttribute(t, response, "fallback_code", testcase.expected)
		})
	}
}

//nolint:forcetypeassert // Fixture mutations target independently defined object shapes.
func TestLocalesDataSourceAliasRetargetingHasNoSnapshotGuarantee(t *testing.T) {
	t.Parallel()
	body := discoveryFixture(t, "locale")
	requests := 0
	// Two different targets could return distinct Locales under the same alias
	// metadata. This fixture demonstrates the observation limit, not CMA timing.
	response := discoveryReadTest(t.Context(), t, NewLocalesDataSource, map[string]any{"space_id": "space", "environment_id": "master"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++

		assert.Equal(t, "/base/spaces/space/environments/master/locales", request.URL.Path)

		if requests == 1 {
			return discoveryHTTPResponse(request, 200, discoveryPage(0, 1, 2, body)), nil
		}

		return discoveryHTTPResponse(request, 200, discoveryPage(1, 1, 2, mutateTestJSON(body, func(document map[string]any) {
			document["sys"].(map[string]any)["id"] = "other-target-locale"
			document["code"] = "fr-FR"
		}))), nil
	}))
	require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
	assert.Equal(t, 2, requests)

	var data LocalesDataSourceModel
	require.False(t, response.State.Get(t.Context(), &data).HasError())
	require.Len(t, data.Locales, 2)
	assert.Equal(t, "master", data.EnvironmentID.ValueString())
	assert.Equal(t, "en-GB", data.Locales[0].Code.ValueString())
	assert.Equal(t, "fr-FR", data.Locales[1].Code.ValueString())
}
