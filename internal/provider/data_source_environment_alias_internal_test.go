package provider

import (
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:forcetypeassert // Fixture mutations target independently defined object shapes.
func TestEnvironmentAliasDataSourceBothTypeSpellings(t *testing.T) {
	t.Parallel()

	body := discoveryFixture(t, "environment_alias")
	for _, spelling := range []string{"Environment Alias", "EnvironmentAlias"} {
		t.Run(spelling, func(t *testing.T) {
			t.Parallel()

			responseBody := mutateTestJSON(body, func(document map[string]any) { document["sys"].(map[string]any)["type"] = spelling })
			client, err := cm.NewClient("https://example.invalid", cm.NewAccessTokenSecuritySource("synthetic"), cm.WithClient(&http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				return discoveryHTTPResponse(request, 200, responseBody), nil
			})}))
			require.NoError(t, err)
			result, err := client.GetEnvironmentAlias(t.Context(), cm.GetEnvironmentAliasParams{SpaceID: "space", EnvironmentAliasID: "item-b"})
			require.NoError(t, err)

			entity, ok := result.(*cm.EnvironmentAlias)
			require.True(t, ok)

			model := NewEnvironmentAliasResourceModelFromResponse(*entity)
			assert.Equal(t, "space/item-b", model.ID.ValueString())
			assert.Equal(t, "item-b", model.EnvironmentAliasID.ValueString())
			assert.Equal(t, "target", model.TargetEnvironmentID.ValueString())
			assert.Equal(t, 9, entity.Sys.Version)
		})
	}
}
