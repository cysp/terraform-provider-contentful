package integration_tests_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppSigningSecretClientTreatsRedactedValueAsOpaqueMetadata(t *testing.T) {
	t.Parallel()

	responseBody := testJSON(map[string]any{
		"sys": map[string]any{
			"type":          "AppSigningSecret",
			"organization":  map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Organization", "id": "organization-id"}},
			"appDefinition": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "AppDefinition", "id": "app-definition-id"}},
		},
		"redactedValue": "changed-redaction-format",
	})

	for name, test := range map[string]struct {
		method string
		status int
	}{
		"read":   {method: http.MethodGet, status: http.StatusOK},
		"create": {method: http.MethodPut, status: http.StatusCreated},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, test.method, r.Method)
				assert.Equal(t, "/organizations/organization-id/app_definitions/app-definition-id/signing_secret", r.URL.Path)

				body, err := io.ReadAll(r.Body)
				assert.NoError(t, err)

				if test.method == http.MethodGet {
					assert.Empty(t, body)
				} else {
					assert.JSONEq(t, testJSON(map[string]any{"value": strings.Repeat("s", 64)}), string(body))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(test.status)
				_, err = io.WriteString(w, responseBody)
				assert.NoError(t, err)
			}))
			t.Cleanup(testServer.Close)

			client := testContentfulManagementClient(t, testServer.URL, cmt.ValidAccessToken)
			if test.method == http.MethodGet {
				response, err := client.GetAppSigningSecret(t.Context(), cm.GetAppSigningSecretParams{OrganizationID: "organization-id", AppDefinitionID: "app-definition-id"})
				require.NoError(t, err)

				secret, ok := response.(*cm.AppSigningSecret)
				require.True(t, ok)
				require.NotNil(t, secret)
				assert.Equal(t, "changed-redaction-format", secret.RedactedValue)

				return
			}

			response, err := client.PutAppSigningSecret(t.Context(), &cm.AppSigningSecretRequestData{Value: strings.Repeat("s", 64)}, cm.PutAppSigningSecretParams{OrganizationID: "organization-id", AppDefinitionID: "app-definition-id"})
			require.NoError(t, err)

			secret, ok := response.(*cm.AppSigningSecretStatusCode)
			require.True(t, ok)
			require.NotNil(t, secret)
			assert.Equal(t, test.status, secret.StatusCode)
			assert.Equal(t, "changed-redaction-format", secret.Response.RedactedValue)
		})
	}
}
