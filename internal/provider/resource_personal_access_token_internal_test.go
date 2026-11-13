package provider

import (
	"bytes"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Direct lifecycle access checks the diagnostic and retained state after decoding
// a successful HTTP response that may not confirm revocation.
func TestPersonalAccessTokenDeleteRequiresRevocationConfirmation(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		responseFields map[string]any
		wantError      bool
	}{
		"omitted timestamp": {
			wantError: true,
		},
		"null timestamp": {
			responseFields: map[string]any{"revokedAt": nil},
			wantError:      true,
		},
		"revocation timestamp": {
			responseFields: map[string]any{"revokedAt": "2026-01-01T00:00:00Z"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			responseBody := map[string]any{
				"sys":    map[string]any{"id": "token-id", "type": "PersonalAccessToken"},
				"name":   "Test token",
				"scopes": []string{"content_management_read"},
			}
			maps.Copy(responseBody, test.responseFields)

			body, err := json.Marshal(responseBody)
			require.NoError(t, err)

			var requests []*http.Request

			client, err := cm.NewClient("https://contentful.invalid", cm.NewAccessTokenSecuritySource("test-token"), cm.WithClient(&http.Client{
				Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
					requests = append(requests, request)

					return &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{"Content-Type": []string{"application/json"}},
						Body:       io.NopCloser(bytes.NewReader(body)),
						Request:    request,
					}, nil
				}),
			}))
			require.NoError(t, err)

			schema := PersonalAccessTokenResourceSchema(t.Context())

			attributes := make(map[string]tftypes.Value, len(schema.Attributes))
			for name, attribute := range schema.Attributes {
				attributes[name] = tftypes.NewValue(attribute.GetType().TerraformType(t.Context()), nil)
			}

			attributes["id"] = tftypes.NewValue(tftypes.String, "token-id")
			state := tfsdk.State{Schema: schema, Raw: tftypes.NewValue(schema.Type().TerraformType(t.Context()), attributes)}
			response := resource.DeleteResponse{State: state}
			implementation := personalAccessTokenResource{providerData: ContentfulProviderData{client: client}}
			implementation.Delete(t.Context(), resource.DeleteRequest{State: state}, &response)

			require.Len(t, requests, 1)
			assert.Equal(t, http.MethodPut, requests[0].Method)
			assert.Equal(t, "/users/me/access_tokens/token-id/revoked", requests[0].URL.Path)

			if test.wantError {
				require.Len(t, response.Diagnostics, 1)
				assert.Equal(t, diag.SeverityError, response.Diagnostics[0].Severity())
				assert.Equal(t, "Failed to revoke personal access token", response.Diagnostics[0].Summary())
				assert.Equal(t, "Contentful returned a token without a revocation timestamp, so the provider could not confirm revocation.", response.Diagnostics[0].Detail())
				assert.True(t, state.Raw.Equal(response.State.Raw), "unconfirmed revocation must retain the resource state")
			} else {
				require.Empty(t, response.Diagnostics)
			}
		})
	}
}
