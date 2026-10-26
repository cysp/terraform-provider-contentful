package provider_test

import (
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mutationEndpointCase struct {
	name                                 string
	schema                               schema.Schema
	plan                                 map[string]any
	body                                 map[string]any
	createMethod, createPath, updatePath string
	createIdentity, updateIdentity       map[string]string
	createID, updateID                   string
	returned                             map[string]any
	versioned                            bool
	generatedIdentity                    string
}

func TestResourceMutationsRetainAddressedEndpointAndReturnedValues(t *testing.T) {
	t.Parallel()

	for _, test := range mutationEndpointCases(t) {
		for _, operation := range []string{"create", "update"} {
			if operation == "update" && test.updatePath == "" {
				continue
			}

			t.Run(test.name+"/"+operation, func(t *testing.T) {
				t.Parallel()

				testMutationEndpoint(t, test, operation)
			})
		}
	}
}

func testMutationEndpoint(t *testing.T, test mutationEndpointCase, operation string) {
	t.Helper()

	method, endpoint, expectedIdentity, expectedID := test.createMethod, test.createPath, test.createIdentity, test.createID
	if operation == "update" {
		method, endpoint, expectedIdentity, expectedID = http.MethodPut, test.updatePath, test.updateIdentity, test.updateID
	}

	body := testJSON(test.body)

	var calls atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		assert.Equal(t, method, r.Method)
		assert.Equal(t, endpoint, r.URL.Path)

		if operation == "update" && test.versioned {
			assert.Equal(t, "31", r.Header.Get("X-Contentful-Version"))
		}

		w.Header().Set("Content-Type", "application/json")

		status := http.StatusOK
		if method == http.MethodPost {
			status = http.StatusCreated
		}

		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	providerServer, err := makeTestAccProtoV6ProviderFactories(WithContentfulURL(server.URL), WithAccessToken("test-token"))["contentful"]()
	require.NoError(t, err)
	providerConfig, err := providerConfigDynamicValue(map[string]any{"url": server.URL, "access_token": "test-token"})
	require.NoError(t, err)
	configured, err := providerServer.ConfigureProvider(t.Context(), &tfprotov6.ConfigureProviderRequest{Config: &providerConfig})
	require.NoError(t, err)
	require.Empty(t, configured.Diagnostics)

	values := maps.Clone(test.plan)
	values["id"] = expectedID
	planned := mutationEndpointDynamicValue(t, test.schema, values)
	prior := nullResourceDynamicValue(t, test.schema)

	if operation == "update" {
		previous := maps.Clone(values)

		previous["name"] = "Prior"
		if _, exists := values["name"]; !exists {
			delete(previous, "name")
		}

		if _, exists := values["value"]; exists {
			previous["value"] = strings.Repeat("b", 64)
		}

		if _, exists := values["parameters"]; exists {
			previous["parameters"] = testJSON(map[string]string{"value": "prior"})
		}

		prior = mutationEndpointDynamicValue(t, test.schema, previous)
	}

	result, err := providerServer.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{
		TypeName: "contentful_" + test.name, PriorState: &prior, PlannedState: &planned, Config: &planned, PlannedPrivate: privateVersionBytes(t, 31),
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, calls.Load(), result.Diagnostics)

	assertMutationEndpoint(t, test, result, expectedIdentity, expectedID)
}

func assertMutationEndpoint(t *testing.T, test mutationEndpointCase, result *tfprotov6.ApplyResourceChangeResponse, expectedIdentity map[string]string, expectedID string) {
	t.Helper()

	for _, diagnostic := range result.Diagnostics {
		t.Logf("%s: %s", diagnostic.Summary, diagnostic.Detail)
	}

	assert.Len(t, result.Diagnostics, len(expectedIdentity), result.Diagnostics)

	for _, diagnostic := range result.Diagnostics {
		assert.Equal(t, tfprotov6.DiagnosticSeverityError, diagnostic.Severity)
		assert.Contains(t, diagnostic.Summary, "identity")
	}

	require.NotNil(t, result.NewState)
	raw, err := result.NewState.Unmarshal(test.schema.Type().TerraformType(t.Context()))
	require.NoError(t, err)

	state := tfsdk.State{Schema: test.schema, Raw: raw}

	var actualID types.String
	require.Empty(t, state.GetAttribute(t.Context(), path.Root("id"), &actualID))
	assert.Equal(t, types.StringValue(expectedID), actualID)
	require.NotNil(t, result.NewIdentity)
	require.NotNil(t, result.NewIdentity.IdentityData)

	identityTypes := make(map[string]tftypes.Type)
	for name := range test.updateIdentity {
		identityTypes[name] = tftypes.String
	}
	// App Key has no remote Update, but key_kid remains a generated identity.
	if test.generatedIdentity != "" {
		identityTypes[test.generatedIdentity] = tftypes.String
	}

	identityRaw, err := result.NewIdentity.IdentityData.Unmarshal(tftypes.Object{AttributeTypes: identityTypes})
	require.NoError(t, err)

	var identity map[string]tftypes.Value
	require.NoError(t, identityRaw.As(&identity))

	for name, expected := range expectedIdentity {
		foundDiagnostic := false

		for _, diagnostic := range result.Diagnostics {
			if diagnostic.Attribute != nil && diagnostic.Attribute.Equal(tftypes.NewAttributePath().WithAttributeName(name)) {
				foundDiagnostic = true
			}
		}

		assert.True(t, foundDiagnostic, "missing attribute diagnostic for %s", name)

		var value types.String
		require.Empty(t, state.GetAttribute(t.Context(), path.Root(name), &value))
		assert.Equal(t, types.StringValue(expected), value, name)

		var identityValue string
		require.NoError(t, identity[name].As(&identityValue))
		assert.Equal(t, expected, identityValue, name)
	}

	var stateAttributes map[string]tftypes.Value
	require.NoError(t, raw.As(&stateAttributes))

	for name, expected := range test.returned {
		if text, ok := expected.(string); ok {
			var actual string
			require.NoError(t, stateAttributes[name].As(&actual))
			assert.Equal(t, text, actual, name)
		} else {
			var actual bool
			require.NoError(t, stateAttributes[name].As(&actual))
			assert.Equal(t, expected, actual, name)
		}
	}

	if test.versioned {
		assertPreviewEnvironmentPrivateVersion(t, result.Private, 73)
	}
}

func mutationEndpointDynamicValue(t *testing.T, resourceSchema schema.Schema, values map[string]any) tfprotov6.DynamicValue {
	t.Helper()
	raw, err := tftypes.ValueFromJSONWithOpts([]byte(testJSON(values)), resourceSchema.Type().TerraformType(t.Context()), tftypes.ValueFromJSONOpts{})
	require.NoError(t, err)
	value, err := tfprotov6.NewDynamicValue(raw.Type(), raw)
	require.NoError(t, err)

	return value
}

func mutationEndpointBody(kind string, fields map[string]any) map[string]any {
	link := func(kind, id string) map[string]any {
		return map[string]any{"sys": map[string]any{"type": "Link", "linkType": kind, "id": id}}
	}
	body := maps.Clone(fields)
	body["sys"] = map[string]any{
		"type": kind, "id": "returned-id", "version": 73, "visibility": "private",
		"space": link("Space", "other-space"), "environment": link("Environment", "other-environment"),
		"organization": link("Organization", "other-organization"), "appDefinition": link("AppDefinition", "other-app"),
		"resourceProvider": link("ResourceProvider", "returned-provider"), "team": link("Team", "returned-team"),
		"createdBy": link("User", "user"), "updatedBy": link("User", "user"), "status": link("Status", "ready"),
	}

	return body
}

func mutationEndpointCases(t *testing.T) []mutationEndpointCase {
	t.Helper()
	ctx := t.Context()
	jwkData := testAccAppKeyJWKFromDER([]byte{1, 2, 3})
	jwk := map[string]any{"alg": "RS256", "kty": "RSA", "use": "sig", "kid": jwkData.kid, "x5c": []string{jwkData.x5c}, "x5t": jwkData.x5t}
	secret := strings.Repeat("a", 64)

	return []mutationEndpointCase{
		{
			name:           "app_definition",
			schema:         AppDefinitionResourceSchema(ctx),
			plan:           map[string]any{"organization_id": "org", "app_definition_id": "app", "name": "Planned", "locations": []any{}},
			body:           mutationEndpointBody("AppDefinition", map[string]any{"name": "Returned", "locations": []any{}}),
			createMethod:   "POST",
			createPath:     "/organizations/org/app_definitions",
			updatePath:     "/organizations/org/app_definitions/app",
			createIdentity: map[string]string{"organization_id": "org"},
			updateIdentity: map[string]string{"organization_id": "org", "app_definition_id": "app"},
			createID:       "org/returned-id",
			updateID:       "org/app",
			returned:       map[string]any{"name": "Returned"},
		},
		{
			name:           "app_installation",
			schema:         AppInstallationResourceSchema(ctx),
			plan:           map[string]any{"space_id": "space", "environment_id": "environment", "app_definition_id": "app", "parameters": testJSON(map[string]string{"value": "planned"})},
			body:           mutationEndpointBody("AppInstallation", map[string]any{"parameters": map[string]string{"value": "returned"}}),
			createMethod:   "PUT",
			createPath:     "/spaces/space/environments/environment/app_installations/app",
			updatePath:     "/spaces/space/environments/environment/app_installations/app",
			createIdentity: map[string]string{"space_id": "space", "environment_id": "environment", "app_definition_id": "app"},
			updateIdentity: map[string]string{"space_id": "space", "environment_id": "environment", "app_definition_id": "app"},
			createID:       "space/environment/app",
			updateID:       "space/environment/app",
			returned:       map[string]any{"parameters": testJSON(map[string]string{"value": "returned"})},
		},
		{
			name:              "app_key",
			generatedIdentity: "key_kid",
			schema:            AppKeyResourceSchema(ctx),
			plan:              map[string]any{"organization_id": "org", "app_definition_id": "app", "jwk": jwk},
			body:              mutationEndpointBody("AppKey", map[string]any{"jwk": jwk}),
			createMethod:      "POST",
			createPath:        "/organizations/org/app_definitions/app/keys",
			createIdentity:    map[string]string{"organization_id": "org", "app_definition_id": "app"},
			updateIdentity:    map[string]string{"organization_id": "org", "app_definition_id": "app"},
			createID:          "org/app/returned-id",
			returned:          map[string]any{"key_kid": "returned-id"},
		},
		{
			name:           "app_signing_secret",
			schema:         AppSigningSecretResourceSchema(ctx),
			plan:           map[string]any{"organization_id": "org", "app_definition_id": "app", "value": secret},
			body:           mutationEndpointBody("AppSigningSecret", map[string]any{"redactedValue": "****aaaa"}),
			createMethod:   "PUT",
			createPath:     "/organizations/org/app_definitions/app/signing_secret",
			updatePath:     "/organizations/org/app_definitions/app/signing_secret",
			createIdentity: map[string]string{"organization_id": "org", "app_definition_id": "app"},
			updateIdentity: map[string]string{"organization_id": "org", "app_definition_id": "app"},
			createID:       "org/app",
			updateID:       "org/app",
			returned:       map[string]any{"value": secret},
		},
		{
			name:           "delivery_api_key",
			schema:         DeliveryAPIKeyResourceSchema(ctx),
			plan:           map[string]any{"space_id": "space", "api_key_id": "key", "name": "Planned", "environments": []any{}},
			body:           mutationEndpointBody("ApiKey", map[string]any{"name": "Returned", "description": "", "accessToken": "returned-token", "environments": []any{}, "preview_api_key": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "PreviewApiKey", "id": "preview"}}}),
			createMethod:   "POST",
			createPath:     "/spaces/space/api_keys",
			updatePath:     "/spaces/space/api_keys/key",
			createIdentity: map[string]string{"space_id": "space"},
			updateIdentity: map[string]string{"space_id": "space", "api_key_id": "key"},
			createID:       "space/returned-id",
			updateID:       "space/key",
			returned:       map[string]any{"name": "Returned", "access_token": "returned-token"},
			versioned:      true,
		},
		{
			name:           "environment",
			schema:         EnvironmentResourceSchema(ctx),
			plan:           map[string]any{"space_id": "space", "environment_id": "environment", "name": "Planned"},
			body:           mutationEndpointBody("Environment", map[string]any{"name": "Returned"}),
			createMethod:   "PUT",
			createPath:     "/spaces/space/environments/environment",
			updatePath:     "/spaces/space/environments/environment",
			createIdentity: map[string]string{"space_id": "space", "environment_id": "environment"},
			updateIdentity: map[string]string{"space_id": "space", "environment_id": "environment"},
			createID:       "space/environment",
			updateID:       "space/environment",
			returned:       map[string]any{"name": "Returned"},
			versioned:      true,
		},
		{
			name:           "environment_alias",
			schema:         EnvironmentAliasResourceSchema(ctx),
			plan:           map[string]any{"space_id": "space", "environment_alias_id": "alias", "target_environment_id": "target"},
			body:           mutationEndpointBody("EnvironmentAlias", map[string]any{"environment": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Environment", "id": "returned-target"}}}),
			createMethod:   "PUT",
			createPath:     "/spaces/space/environment_aliases/alias",
			updatePath:     "/spaces/space/environment_aliases/alias",
			createIdentity: map[string]string{"space_id": "space", "environment_alias_id": "alias"},
			updateIdentity: map[string]string{"space_id": "space", "environment_alias_id": "alias"},
			createID:       "space/alias",
			updateID:       "space/alias",
			returned:       map[string]any{"target_environment_id": "returned-target"},
			versioned:      true,
		},
		{
			name:           "extension",
			schema:         ExtensionResourceSchema(ctx),
			plan:           map[string]any{"space_id": "space", "environment_id": "environment", "extension_id": "extension", "extension": map[string]any{"name": "Planned", "srcdoc": "<p>planned</p>", "field_types": []any{}, "sidebar": false}},
			body:           mutationEndpointBody("Extension", map[string]any{"extension": map[string]any{"name": "Returned", "srcdoc": "<p>returned</p>", "fieldTypes": []any{}, "sidebar": false}}),
			createMethod:   "PUT",
			createPath:     "/spaces/space/environments/environment/extensions/extension",
			updatePath:     "/spaces/space/environments/environment/extensions/extension",
			createIdentity: map[string]string{"space_id": "space", "environment_id": "environment", "extension_id": "extension"},
			updateIdentity: map[string]string{"space_id": "space", "environment_id": "environment", "extension_id": "extension"},
			createID:       "space/environment/extension",
			updateID:       "space/environment/extension",
			versioned:      true,
		},
		{
			name:           "resource_provider",
			schema:         ResourceProviderResourceSchema(ctx),
			plan:           map[string]any{"organization_id": "org", "app_definition_id": "app", "resource_provider_id": "provider", "function_id": "planned-function"},
			body:           mutationEndpointBody("ResourceProvider", map[string]any{"type": "function", "function": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Function", "id": "returned-function"}}}),
			createMethod:   "PUT",
			createPath:     "/organizations/org/app_definitions/app/resource_provider",
			updatePath:     "/organizations/org/app_definitions/app/resource_provider",
			createIdentity: map[string]string{"organization_id": "org", "app_definition_id": "app"},
			updateIdentity: map[string]string{"organization_id": "org", "app_definition_id": "app"},
			createID:       "org/app",
			updateID:       "org/app",
			returned:       map[string]any{"resource_provider_id": "returned-id", "function_id": "returned-function"},
		},
		{
			name:           "resource_type",
			schema:         ResourceTypeResourceSchema(ctx),
			plan:           map[string]any{"organization_id": "org", "app_definition_id": "app", "resource_type_id": "provider:type", "resource_provider_id": "provider", "name": "Planned", "default_field_mapping": map[string]any{"title": "title"}},
			body:           mutationEndpointBody("ResourceType", map[string]any{"name": "Returned", "defaultFieldMapping": map[string]any{"title": "returned-title"}}),
			createMethod:   "PUT",
			createPath:     "/organizations/org/app_definitions/app/resource_provider/resource_types/provider:type",
			updatePath:     "/organizations/org/app_definitions/app/resource_provider/resource_types/provider:type",
			createIdentity: map[string]string{"organization_id": "org", "app_definition_id": "app", "resource_type_id": "provider:type"},
			updateIdentity: map[string]string{"organization_id": "org", "app_definition_id": "app", "resource_type_id": "provider:type"},
			createID:       "org/app/provider:type",
			updateID:       "org/app/provider:type",
			returned:       map[string]any{"name": "Returned", "resource_provider_id": "returned-provider"},
		},
		{
			name:           "space_enablements",
			schema:         SpaceEnablementsResourceSchema(ctx),
			plan:           map[string]any{"space_id": "space", "cross_space_links": false, "space_templates": false},
			body:           mutationEndpointBody("SpaceEnablement", map[string]any{"crossSpaceLinks": map[string]any{"enabled": true}, "spaceTemplates": map[string]any{"enabled": true}}),
			createMethod:   "PUT",
			createPath:     "/spaces/space/enablements",
			updatePath:     "/spaces/space/enablements",
			createIdentity: map[string]string{"space_id": "space"},
			updateIdentity: map[string]string{"space_id": "space"},
			createID:       "space",
			updateID:       "space",
			returned:       map[string]any{"cross_space_links": true},
			versioned:      true,
		},
		{
			name:           "tag",
			schema:         TagResourceSchema(ctx),
			plan:           map[string]any{"space_id": "space", "environment_id": "environment", "tag_id": "tag", "name": "Planned", "visibility": "private"},
			body:           mutationEndpointBody("Tag", map[string]any{"name": "Returned"}),
			createMethod:   "PUT",
			createPath:     "/spaces/space/environments/environment/tags/tag",
			updatePath:     "/spaces/space/environments/environment/tags/tag",
			createIdentity: map[string]string{"space_id": "space", "environment_id": "environment", "tag_id": "tag"},
			updateIdentity: map[string]string{"space_id": "space", "environment_id": "environment", "tag_id": "tag"},
			createID:       "space/environment/tag",
			updateID:       "space/environment/tag",
			returned:       map[string]any{"name": "Returned"},
			versioned:      true,
		},
		{
			name:           "team",
			schema:         TeamResourceSchema(ctx),
			plan:           map[string]any{"organization_id": "org", "team_id": "team", "name": "Planned"},
			body:           mutationEndpointBody("Team", map[string]any{"name": "Returned", "description": ""}),
			createMethod:   "POST",
			createPath:     "/organizations/org/teams",
			updatePath:     "/organizations/org/teams/team",
			createIdentity: map[string]string{"organization_id": "org"},
			updateIdentity: map[string]string{"organization_id": "org", "team_id": "team"},
			createID:       "org/returned-id",
			updateID:       "org/team",
			returned:       map[string]any{"name": "Returned"},
			versioned:      true,
		},
		{
			name:           "team_space_membership",
			schema:         TeamSpaceMembershipResourceSchema(ctx),
			plan:           map[string]any{"space_id": "space", "team_space_membership_id": "membership", "team_id": "team", "admin": false, "roles": []any{}},
			body:           mutationEndpointBody("TeamSpaceMembership", map[string]any{"admin": true, "roles": []any{}}),
			createMethod:   "POST",
			createPath:     "/spaces/space/team_space_memberships",
			updatePath:     "/spaces/space/team_space_memberships/membership",
			createIdentity: map[string]string{"space_id": "space"},
			updateIdentity: map[string]string{"space_id": "space", "team_space_membership_id": "membership"},
			createID:       "space/returned-id",
			updateID:       "space/membership",
			returned:       map[string]any{"team_id": "returned-team", "admin": true},
		},
	}
}
