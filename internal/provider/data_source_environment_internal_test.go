package provider

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:forcetypeassert // Fixture mutations target independently defined object shapes.
func TestEnvironmentDataSourceStatusesAndSharedProjection(t *testing.T) {
	t.Parallel()

	body := discoveryFixture(t, "environment")
	for _, status := range []string{"queued", "inProgress", "ready", "failed", "future-status"} {
		t.Run(status, func(t *testing.T) {
			t.Parallel()

			count := 0
			response := discoveryReadTest(t.Context(), t, NewEnvironmentDataSource, map[string]any{"space_id": "space", "environment_id": "item-b"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
				count++

				return discoveryHTTPResponse(request, 200, mutateTestJSON(body, func(document map[string]any) {
					document["sys"].(map[string]any)["status"].(map[string]any)["sys"].(map[string]any)["id"] = status
				})), nil
			}))
			require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
			assert.Equal(t, 1, count)
			discoveryStateAttribute(t, response, "status", fmt.Sprintf("%q", status))
		})
	}

	for _, aliased := range []bool{false, true} {
		t.Run(strconv.FormatBool(aliased), func(t *testing.T) {
			t.Parallel()

			responseBody := body
			if !aliased {
				responseBody = mutateTestJSON(body, func(document map[string]any) { delete(document["sys"].(map[string]any), "aliasedEnvironment") })
			}

			client, err := cm.NewClient("https://example.invalid", cm.NewAccessTokenSecuritySource("synthetic"), cm.WithClient(&http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				return discoveryHTTPResponse(request, 200, responseBody), nil
			})}))
			require.NoError(t, err)
			result, err := client.GetEnvironment(t.Context(), cm.GetEnvironmentParams{SpaceID: "space", EnvironmentID: "item-b"})
			require.NoError(t, err)

			entity, ok := result.(*cm.Environment)
			require.True(t, ok)

			resource := NewEnvironmentResourceModelFromResponse(*entity)
			waiter := NewEnvironmentStatusReadyModelFromResponse(*entity)

			assert.Equal(t, "space/item-b", resource.ID.ValueString())
			assert.Equal(t, "Second", resource.Name.ValueString())
			assert.Equal(t, "queued", resource.Status.ValueString())
			assert.Equal(t, "space/item-b", waiter.ID.ValueString())
			assert.Equal(t, "queued", waiter.Status.ValueString())
			assert.Equal(t, 7, entity.Sys.Version)
		})
	}
}

func TestEnvironmentDataSourceReducedPublishedExampleFails(t *testing.T) {
	t.Parallel()

	for _, body := range []string{
		testJSON(map[string]any{
			"sys": map[string]any{
				"id":      "item-b",
				"type":    "Environment",
				"version": 1,
				"space":   map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Space", "id": "space"}},
			},
		}),
		testJSON(map[string]any{
			"sys": map[string]any{
				"id":      "item-b",
				"type":    "Environment",
				"version": 1,
				"space":   map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Space", "id": "space"}},
				"status":  map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Status", "id": "ready"}},
			},
		}),
	} {
		response := discoveryReadTest(t.Context(), t, NewEnvironmentDataSource, map[string]any{"space_id": "space", "environment_id": "item-b"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
			return discoveryHTTPResponse(request, 200, body), nil
		}))
		assert.True(t, response.Diagnostics.HasError())
		assert.True(t, response.State.Raw.IsNull())
	}
}
