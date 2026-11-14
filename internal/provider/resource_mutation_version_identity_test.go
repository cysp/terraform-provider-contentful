package provider_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The existing request fixtures supply valid Terraform models; the response
// fixtures contain independently written contradictory identities and version 73.
func TestVersionedMutationsRejectVersionFromContradictoryIdentity(t *testing.T) {
	t.Parallel()

	endpoints := map[string]string{
		"contentful_live_preview_variables": "/spaces/space/environments/environment/live_preview/variables",
		"contentful_content_type":           "/spaces/space/environments/environment/content_types/content-type",
		"contentful_entry":                  "/spaces/space/environments/environment/entries/entry",
		"contentful_preview_environment":    "/spaces/space/preview_environments/preview",
		"contentful_delivery_api_key":       "/spaces/space/api_keys/key",
		"contentful_webhook":                "/spaces/space/webhook_definitions/webhook",
		"contentful_extension":              "/spaces/space/environments/environment/extensions/extension",
		"contentful_space_enablements":      "/spaces/space/enablements",
		"contentful_role":                   "/spaces/space/roles/role",
		"contentful_team":                   "/organizations/organization/teams/team",
		"contentful_environment":            "/spaces/space/environments/environment",
		"contentful_environment_alias":      "/spaces/space/environment_aliases/alias",
		"contentful_editor_interface":       "/spaces/space/environments/environment/content_types/content-type/editor_interface",
		"contentful_locale":                 "/spaces/space/environments/environment/locales/locale",
		"contentful_tag":                    "/spaces/space/environments/environment/tags/tag",
	}

	bodies := make(map[string]map[string]any)
	for _, fixture := range readEndpointCases(t) {
		bodies["contentful_"+fixture.name] = fixture.body
	}

	for _, test := range requiredPrivateVersionResourceCases(t) {
		if test.delete {
			continue
		}

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			body, found := bodies[test.typeName]
			require.True(t, found)
			endpoint, found := endpoints[test.typeName]
			require.True(t, found)

			var requests atomic.Int64

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Equal(t, endpoint, r.URL.Path)
				assert.Equal(t, "31", r.Header.Get("X-Contentful-Version"))
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(testJSON(body)))
			}))
			t.Cleanup(server.Close)
			protocol, err := makeTestAccProtoV6ProviderFactories(WithContentfulURL(server.URL), WithAccessToken("test-token"))["contentful"]()
			require.NoError(t, err)
			configuration, err := providerConfigDynamicValue(map[string]any{"url": server.URL, "access_token": "test-token"})
			require.NoError(t, err)
			configured, err := protocol.ConfigureProvider(t.Context(), &tfprotov6.ConfigureProviderRequest{Config: &configuration})
			require.NoError(t, err)
			require.Empty(t, configured.Diagnostics)
			prior := resourceModelDynamicValue(t, test.resourceSchema, test.model)

			plannedModel := test.model
			if test.plannedModel != nil {
				plannedModel = test.plannedModel
			}

			plan := resourceModelDynamicValue(t, test.resourceSchema, plannedModel)
			response, err := protocol.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{TypeName: test.typeName, PriorState: &prior, PlannedState: &plan, Config: &plan, PlannedPrivate: privateVersionBytes(t, 31)})
			require.NoError(t, err)
			require.EqualValues(t, 1, requests.Load(), response.Diagnostics)
			require.NotEmpty(t, response.Diagnostics)

			var private map[string][]byte
			if len(response.Private) != 0 {
				require.NoError(t, json.Unmarshal(response.Private, &private))
			}

			assert.Empty(t, private["version"], "unrelated response version 73 must not replace target version 31")
			assert.Empty(t, private["pending_publication_version"])
			assert.Empty(t, private["pending_activation_version"])
			require.NotNil(t, response.NewState)
			require.NotNil(t, response.NewIdentity)
			// Retry from the returned recovery state with refresh disabled. This checks
			// the actual caller, not just the version-publication helper.
			retry, err := protocol.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{TypeName: test.typeName, PriorState: response.NewState, PlannedState: &plan, Config: &plan, PlannedPrivate: response.Private})
			require.NoError(t, err)
			require.EqualValues(t, 1, requests.Load(), "an unconfirmed response cannot authorize another write")

			var summaries []string
			for _, diagnostic := range retry.Diagnostics {
				summaries = append(summaries, diagnostic.Summary)
			}

			assert.Contains(t, strings.Join(summaries, "; "), "Private version is unavailable")
		})
	}
}
