package provider_test

import (
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type readEndpointCase struct {
	name             string
	schema           schema.Schema
	prior, body      map[string]any
	endpoint, id     string
	pendingKey       string
	diagnosticDetail string
}

func TestResourceReadsRejectContradictoryEndpointBeforePublication(t *testing.T) {
	t.Parallel()

	for _, test := range readEndpointCases(t) {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var requests atomic.Int64

			body := testJSON(test.body)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, test.endpoint, r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(body))
			}))
			t.Cleanup(server.Close)
			providerServer, err := makeTestAccProtoV6ProviderFactories(WithContentfulURL(server.URL), WithAccessToken("test-token"))["contentful"]()
			require.NoError(t, err)
			config, err := providerConfigDynamicValue(map[string]any{"url": server.URL, "access_token": "test-token"})
			require.NoError(t, err)
			configured, err := providerServer.ConfigureProvider(t.Context(), &tfprotov6.ConfigureProviderRequest{Config: &config})
			require.NoError(t, err)
			require.Empty(t, configured.Diagnostics)

			values := maps.Clone(test.prior)
			values["id"] = test.id
			valueType := test.schema.Type().TerraformType(t.Context())
			before, err := tftypes.ValueFromJSONWithOpts([]byte(testJSON(values)), valueType, tftypes.ValueFromJSONOpts{})
			require.NoError(t, err)
			current, err := tfprotov6.NewDynamicValue(valueType, before)
			require.NoError(t, err)

			private := privateVersionBytes(t, 31)
			if test.pendingKey != "" {
				private = []byte(testJSON(map[string][]byte{"version": []byte("31"), test.pendingKey: []byte("31")}))
			}

			result, err := providerServer.ReadResource(t.Context(), &tfprotov6.ReadResourceRequest{TypeName: "contentful_" + test.name, CurrentState: &current, Private: private})
			require.NoError(t, err)
			assert.EqualValues(t, 1, requests.Load())
			assert.NotEmpty(t, result.Diagnostics)

			for _, diagnostic := range result.Diagnostics {
				assert.Equal(t, tfprotov6.DiagnosticSeverityError, diagnostic.Severity)
				assert.Contains(t, diagnostic.Summary, "identity")

				if test.diagnosticDetail != "" {
					assert.Equal(t, test.diagnosticDetail, diagnostic.Detail)
				}
			}

			require.NotNil(t, result.NewState)
			after, err := result.NewState.Unmarshal(valueType)
			require.NoError(t, err)
			assert.True(t, before.Equal(after), "prior %s, returned %s", before, after)
			assert.JSONEq(t, string(private), string(result.Private))
		})
	}
}

func readEndpointBody(kind string, fields map[string]any) map[string]any {
	link := func(kind, id string) map[string]any {
		return map[string]any{"sys": map[string]any{"type": "Link", "linkType": kind, "id": id}}
	}
	body := maps.Clone(fields)
	body["sys"] = map[string]any{
		"type": kind, "id": "returned-id", "version": 73, "visibility": "private",
		"space": link("Space", "other-space"), "environment": link("Environment", "other-environment"),
		"contentType":  link("ContentType", "other-type"),
		"organization": link("Organization", "other-organization"), "appDefinition": link("AppDefinition", "other-app"),
		"resourceProvider": link("ResourceProvider", "returned-provider"), "team": link("Team", "returned-team"),
		"createdBy": link("User", "user"), "updatedBy": link("User", "user"), "status": link("Status", "ready"),
	}

	return body
}

//nolint:maintidx // One literal table independently covers every registered managed Read endpoint.
func readEndpointCases(t *testing.T) []readEndpointCase {
	t.Helper()
	ctx := t.Context()
	jwkData := testAccAppKeyJWKFromDER([]byte{1, 2, 3})
	jwk := map[string]any{"alg": "RS256", "kty": "RSA", "use": "sig", "kid": jwkData.kid, "x5c": []string{jwkData.x5c}, "x5t": jwkData.x5t}
	secret := strings.Repeat("a", 64)

	return []readEndpointCase{
		{
			name:     "app_action",
			schema:   AppActionResourceSchema(ctx),
			prior:    map[string]any{"organization_id": "org", "app_definition_id": "app", "app_action_id": "action", "name": "Prior", "category": "Custom", "type": "endpoint"},
			body:     readEndpointBody("AppAction", map[string]any{"name": "Returned", "category": "Custom", "type": "endpoint"}),
			endpoint: "/organizations/org/app_definitions/app/actions/action",
			id:       "org/app/action",
		},
		{
			name:             "app_event_subscription",
			diagnosticDetail: "The response parent differs from the addressed singleton. Review the subscription before applying again.",
			schema:           AppEventSubscriptionResourceSchema(ctx),
			prior:            map[string]any{"organization_id": "org", "app_definition_id": "app", "topics": []any{"Entry.create"}},
			body:             readEndpointBody("AppEventSubscription", map[string]any{"topics": []any{"Entry.delete"}}),
			endpoint:         "/organizations/org/app_definitions/app/event_subscription",
			id:               "org/app",
		},
		{
			name:       "content_type",
			schema:     ContentTypeResourceSchema(ctx),
			prior:      map[string]any{"space_id": "space", "environment_id": "environment", "content_type_id": "type", "name": "Prior", "fields": []any{}},
			body:       readEndpointBody("ContentType", map[string]any{"name": "Returned", "displayField": nil, "fields": []any{}}),
			endpoint:   "/spaces/space/environments/environment/content_types/type",
			id:         "space/environment/type",
			pendingKey: "pending_activation_version",
		},
		{
			name:       "entry",
			schema:     EntryResourceSchema(ctx),
			prior:      map[string]any{"space_id": "space", "environment_id": "environment", "entry_id": "entry", "content_type_id": "type", "fields": map[string]any{}},
			body:       readEndpointBody("Entry", map[string]any{"fields": map[string]any{}}),
			endpoint:   "/spaces/space/environments/environment/entries/entry",
			id:         "space/environment/entry",
			pendingKey: "pending_publication_version",
		},
		{
			name:     "editor_interface",
			schema:   EditorInterfaceResourceSchema(ctx),
			prior:    map[string]any{"space_id": "space", "environment_id": "environment", "content_type_id": "type", "controls": []any{}},
			body:     readEndpointBody("EditorInterface", map[string]any{"controls": []any{}}),
			endpoint: "/spaces/space/environments/environment/content_types/type/editor_interface",
			id:       "space/environment/type",
		},
		{
			name:     "locale",
			schema:   LocaleResourceSchema(ctx),
			prior:    map[string]any{"space_id": "space", "environment_id": "environment", "locale_id": "locale", "name": "Prior", "code": "en-US"},
			body:     readEndpointBody("Locale", map[string]any{"name": "Returned", "code": "en-US", "fallbackCode": nil, "default": false, "contentDeliveryApi": true, "contentManagementApi": true, "optional": true}),
			endpoint: "/spaces/space/environments/environment/locales/locale",
			id:       "space/environment/locale",
		},
		{
			name:     "personal_access_token",
			schema:   PersonalAccessTokenResourceSchema(ctx),
			prior:    map[string]any{"name": "Prior", "scopes": []any{"content_management_manage"}},
			body:     readEndpointBody("PersonalAccessToken", map[string]any{"name": "Returned", "scopes": []any{"content_management_manage"}}),
			endpoint: "/users/me/access_tokens/token",
			id:       "token",
		},
		{
			name:     "preview_environment",
			schema:   PreviewEnvironmentResourceSchema(ctx),
			prior:    map[string]any{"space_id": "space", "preview_environment_id": "preview", "name": "Prior", "description": ""},
			body:     readEndpointBody("PreviewEnvironment", map[string]any{"name": "Returned", "description": "", "configurations": []any{}}),
			endpoint: "/spaces/space/preview_environments/preview",
			id:       "space/preview",
		},
		{
			name:     "live_preview_variables",
			schema:   LivePreviewVariablesResourceSchema(ctx),
			prior:    map[string]any{"space_id": "space", "environment_id": "environment", "variables": testJSON(map[string]any{"prior": "value"})},
			body:     readEndpointBody("LivePreviewVariables", map[string]any{"variables": map[string]any{"returned": "value"}}),
			endpoint: "/spaces/space/environments/environment/live_preview/variables",
			id:       "space/environment",
		},
		{
			name:     "role",
			schema:   RoleResourceSchema(ctx),
			prior:    map[string]any{"space_id": "space", "role_id": "role", "name": "Prior", "permissions": map[string]any{}},
			body:     readEndpointBody("Role", map[string]any{"name": "Returned", "permissions": map[string]any{}}),
			endpoint: "/spaces/space/roles/role",
			id:       "space/role",
		},
		{
			name:     "taxonomy_concept",
			schema:   TaxonomyConceptResourceSchema(ctx),
			prior:    map[string]any{"organization_id": "org", "concept_id": "concept", "pref_label": map[string]any{"en-US": "Prior"}},
			body:     readEndpointBody("TaxonomyConcept", map[string]any{"prefLabel": map[string]any{"en-US": "Returned"}, "conceptSchemes": []any{}}),
			endpoint: "/organizations/org/taxonomy/concepts/concept",
			id:       "org/concept",
		},
		{
			name:     "taxonomy_concept_scheme",
			schema:   TaxonomyConceptSchemeResourceSchema(ctx),
			prior:    map[string]any{"organization_id": "org", "concept_scheme_id": "scheme", "pref_label": map[string]any{"en-US": "Prior"}},
			body:     readEndpointBody("TaxonomyConceptScheme", map[string]any{"prefLabel": map[string]any{"en-US": "Returned"}, "topConcepts": []any{}, "concepts": []any{}, "totalConcepts": 0}),
			endpoint: "/organizations/org/taxonomy/concept-schemes/scheme",
			id:       "org/scheme",
		},
		{
			name:     "webhook",
			schema:   WebhookResourceSchema(ctx),
			prior:    map[string]any{"space_id": "space", "webhook_id": "webhook", "name": "Prior", "url": "https://example.test/prior", "topics": []any{"Entry.create"}},
			body:     readEndpointBody("WebhookDefinition", map[string]any{"name": "Returned", "url": "https://example.test/returned", "topics": []any{"Entry.delete"}}),
			endpoint: "/spaces/space/webhook_definitions/webhook",
			id:       "space/webhook",
		},
		{
			name:     "webhook_signing_secret",
			schema:   WebhookSigningSecretResourceSchema(ctx),
			prior:    map[string]any{"space_id": "space", "value": secret},
			body:     readEndpointBody("WebhookSigningSecret", map[string]any{"redactedValue": "****"}),
			endpoint: "/spaces/space/webhook_settings/signing_secret",
			id:       "space",
		},

		{
			name:     "app_definition",
			schema:   AppDefinitionResourceSchema(ctx),
			prior:    map[string]any{"organization_id": "org", "app_definition_id": "app", "name": "Planned", "locations": []any{}},
			body:     readEndpointBody("AppDefinition", map[string]any{"name": "Returned", "locations": []any{}}),
			endpoint: "/organizations/org/app_definitions/app",
			id:       "org/app",
		},
		{
			name:     "app_installation",
			schema:   AppInstallationResourceSchema(ctx),
			prior:    map[string]any{"space_id": "space", "environment_id": "environment", "app_definition_id": "app", "parameters": testJSON(map[string]string{"value": "planned"})},
			body:     readEndpointBody("AppInstallation", map[string]any{"parameters": map[string]string{"value": "returned"}}),
			endpoint: "/spaces/space/environments/environment/app_installations/app",
			id:       "space/environment/app",
		},
		{
			name:     "app_key",
			endpoint: "/organizations/org/app_definitions/app/keys/key",
			id:       "org/app/key",
			schema:   AppKeyResourceSchema(ctx),
			prior:    map[string]any{"organization_id": "org", "app_definition_id": "app", "jwk": jwk, "key_kid": "key"},
			body:     readEndpointBody("AppKey", map[string]any{"jwk": jwk}),
		},
		{
			name:     "app_signing_secret",
			schema:   AppSigningSecretResourceSchema(ctx),
			prior:    map[string]any{"organization_id": "org", "app_definition_id": "app", "value": secret},
			body:     readEndpointBody("AppSigningSecret", map[string]any{"redactedValue": "****aaaa"}),
			endpoint: "/organizations/org/app_definitions/app/signing_secret",
			id:       "org/app",
		},
		{
			name:     "delivery_api_key",
			schema:   DeliveryAPIKeyResourceSchema(ctx),
			prior:    map[string]any{"space_id": "space", "api_key_id": "key", "name": "Planned", "environments": []any{}},
			body:     readEndpointBody("ApiKey", map[string]any{"name": "Returned", "description": "", "accessToken": "returned-token", "environments": []any{}, "preview_api_key": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "PreviewApiKey", "id": "preview"}}}),
			endpoint: "/spaces/space/api_keys/key",
			id:       "space/key",
		},
		{
			name:     "environment",
			schema:   EnvironmentResourceSchema(ctx),
			prior:    map[string]any{"space_id": "space", "environment_id": "environment", "name": "Planned"},
			body:     readEndpointBody("Environment", map[string]any{"name": "Returned"}),
			endpoint: "/spaces/space/environments/environment",
			id:       "space/environment",
		},
		{
			name:     "environment_alias",
			schema:   EnvironmentAliasResourceSchema(ctx),
			prior:    map[string]any{"space_id": "space", "environment_alias_id": "alias", "target_environment_id": "target"},
			body:     readEndpointBody("EnvironmentAlias", map[string]any{"environment": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Environment", "id": "returned-target"}}}),
			endpoint: "/spaces/space/environment_aliases/alias",
			id:       "space/alias",
		},
		{
			name:     "extension",
			schema:   ExtensionResourceSchema(ctx),
			prior:    map[string]any{"space_id": "space", "environment_id": "environment", "extension_id": "extension", "extension": map[string]any{"name": "Planned", "srcdoc": "<p>planned</p>", "field_types": []any{}, "sidebar": false}},
			body:     readEndpointBody("Extension", map[string]any{"extension": map[string]any{"name": "Returned", "srcdoc": "<p>returned</p>", "fieldTypes": []any{}, "sidebar": false}}),
			endpoint: "/spaces/space/environments/environment/extensions/extension",
			id:       "space/environment/extension",
		},
		{
			name:     "resource_provider",
			schema:   ResourceProviderResourceSchema(ctx),
			prior:    map[string]any{"organization_id": "org", "app_definition_id": "app", "resource_provider_id": "provider", "function_id": "planned-function"},
			body:     readEndpointBody("ResourceProvider", map[string]any{"type": "function", "function": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Function", "id": "returned-function"}}}),
			endpoint: "/organizations/org/app_definitions/app/resource_provider",
			id:       "org/app",
		},
		{
			name:     "resource_type",
			schema:   ResourceTypeResourceSchema(ctx),
			prior:    map[string]any{"organization_id": "org", "app_definition_id": "app", "resource_type_id": "provider:type", "resource_provider_id": "provider", "name": "Planned", "default_field_mapping": map[string]any{"title": "title"}},
			body:     readEndpointBody("ResourceType", map[string]any{"name": "Returned", "defaultFieldMapping": map[string]any{"title": "returned-title"}}),
			endpoint: "/organizations/org/app_definitions/app/resource_provider/resource_types/provider:type",
			id:       "org/app/provider:type",
		},
		{
			name:     "space_enablements",
			schema:   SpaceEnablementsResourceSchema(ctx),
			prior:    map[string]any{"space_id": "space", "cross_space_links": false, "space_templates": false},
			body:     readEndpointBody("SpaceEnablement", map[string]any{"crossSpaceLinks": map[string]any{"enabled": true}, "spaceTemplates": map[string]any{"enabled": true}}),
			endpoint: "/spaces/space/enablements",
			id:       "space",
		},
		{
			name:     "tag",
			schema:   TagResourceSchema(ctx),
			prior:    map[string]any{"space_id": "space", "environment_id": "environment", "tag_id": "tag", "name": "Planned", "visibility": "private"},
			body:     readEndpointBody("Tag", map[string]any{"name": "Returned"}),
			endpoint: "/spaces/space/environments/environment/tags/tag",
			id:       "space/environment/tag",
		},
		{
			name:     "team",
			schema:   TeamResourceSchema(ctx),
			prior:    map[string]any{"organization_id": "org", "team_id": "team", "name": "Planned"},
			body:     readEndpointBody("Team", map[string]any{"name": "Returned", "description": ""}),
			endpoint: "/organizations/org/teams/team",
			id:       "org/team",
		},
		{
			name:     "team_space_membership",
			schema:   TeamSpaceMembershipResourceSchema(ctx),
			prior:    map[string]any{"space_id": "space", "team_space_membership_id": "membership", "team_id": "team", "admin": false, "roles": []any{}},
			body:     readEndpointBody("TeamSpaceMembership", map[string]any{"admin": true, "roles": []any{}}),
			endpoint: "/spaces/space/team_space_memberships/membership",
			id:       "space/membership",
		},
	}
}
