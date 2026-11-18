package provider_test

import (
	"encoding/json"
	"maps"
	"net/http"
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

// Each mutation supplies its own valid prior state, planned changes, and an
// independently authored contradictory response with version 73.
func TestVersionedMutationsRejectVersionFromContradictoryIdentity(t *testing.T) {
	t.Parallel()

	for _, test := range versionedMutationIdentityCases(t) {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var requests atomic.Int64

			protocol := configuredProtocolServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				assert.Equal(t, test.method, r.Method)
				assert.Equal(t, test.endpoint, r.URL.Path)
				assert.Equal(t, "31", r.Header.Get("X-Contentful-Version"))
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(testJSON(test.body)))
			}))
			prior := mutationEndpointDynamicValue(t, test.schema, test.prior)
			planned := maps.Clone(test.prior)
			maps.Copy(planned, test.changes)
			plan := mutationEndpointDynamicValue(t, test.schema, planned)
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

type versionedMutationIdentityCase struct {
	name, typeName, method, endpoint string
	schema                           schema.Schema
	prior, changes, body             map[string]any
}

func versionedMutationIdentityCases(t *testing.T) []versionedMutationIdentityCase {
	t.Helper()
	ctx := t.Context()

	return []versionedMutationIdentityCase{
		{
			name: "live preview variables update", typeName: "contentful_live_preview_variables", method: http.MethodPut,
			endpoint: "/spaces/space/environments/environment/live_preview/variables", schema: LivePreviewVariablesResourceSchema(ctx),
			prior:   map[string]any{"id": "space/environment", "space_id": "space", "environment_id": "environment", "variables": testJSON(map[string]any{})},
			changes: map[string]any{"variables": testJSON(map[string]any{"updated": "value"})},
			body:    readEndpointBody("LivePreviewVariables", map[string]any{"variables": map[string]any{"returned": "value"}}),
		},
		{
			name: "content type update", typeName: "contentful_content_type", method: http.MethodPut,
			endpoint: "/spaces/space/environments/environment/content_types/content-type", schema: ContentTypeResourceSchema(ctx),
			prior:   map[string]any{"id": "space/environment/content-type", "space_id": "space", "environment_id": "environment", "content_type_id": "content-type", "name": "Prior", "description": "A test content type", "display_field": "title", "fields": []any{map[string]any{"id": "title", "name": "Title", "type": "Symbol", "required": true, "localized": false, "disabled": false, "omitted": false, "validations": []any{}}}, "published_version": 1},
			changes: map[string]any{"name": "Planned"},
			body:    readEndpointBody("ContentType", map[string]any{"name": "Returned", "displayField": nil, "fields": []any{}}),
		},
		{
			name: "entry update", typeName: "contentful_entry", method: http.MethodPut,
			endpoint: "/spaces/space/environments/environment/entries/entry", schema: EntryResourceSchema(ctx),
			prior:   map[string]any{"id": "space/environment/entry", "space_id": "space", "environment_id": "environment", "entry_id": "entry", "content_type_id": "content-type", "fields": map[string]any{}, "metadata": map[string]any{"concepts": []any{}, "tags": []any{}}},
			changes: map[string]any{"fields": map[string]any{"managed": map[string]any{"en-US": testJSON("changed")}}},
			body:    readEndpointBody("Entry", map[string]any{"fields": map[string]any{}}),
		},
		{
			name: "preview environment update", typeName: "contentful_preview_environment", method: http.MethodPut,
			endpoint: "/spaces/space/preview_environments/preview", schema: PreviewEnvironmentResourceSchema(ctx),
			prior:   map[string]any{"id": "space/preview", "space_id": "space", "preview_environment_id": "preview", "name": "Prior", "description": "", "content_type_configurations": map[string]any{"page": map[string]any{"url": "https://preview.invalid/page"}}},
			changes: map[string]any{"name": "Planned"},
			body:    readEndpointBody("PreviewEnvironment", map[string]any{"name": "Returned", "description": "", "configurations": []any{}}),
		},
		{
			name: "delivery API key update", typeName: "contentful_delivery_api_key", method: http.MethodPut,
			endpoint: "/spaces/space/api_keys/key", schema: DeliveryAPIKeyResourceSchema(ctx),
			prior: map[string]any{"id": "space/key", "space_id": "space", "api_key_id": "key", "name": "Key", "environments": []any{}, "access_token": "redacted", "preview_api_key_id": "preview"},
			body:  readEndpointBody("ApiKey", map[string]any{"name": "Returned", "description": "", "accessToken": "returned-token", "environments": []any{}, "preview_api_key": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "PreviewApiKey", "id": "preview"}}}),
		},
		{
			name: "webhook update", typeName: "contentful_webhook", method: http.MethodPut,
			endpoint: "/spaces/space/webhook_definitions/webhook", schema: WebhookResourceSchema(ctx),
			prior: map[string]any{"id": "space/webhook", "space_id": "space", "webhook_id": "webhook", "name": "Prior", "url": "https://example.com/webhook", "active": true},
			body:  readEndpointBody("WebhookDefinition", map[string]any{"name": "Returned", "url": "https://example.test/returned", "topics": []any{"Entry.delete"}}),
		},
		{
			name: "extension update", typeName: "contentful_extension", method: http.MethodPut,
			endpoint: "/spaces/space/environments/environment/extensions/extension", schema: ExtensionResourceSchema(ctx),
			prior: map[string]any{"id": "space/environment/extension", "space_id": "space", "environment_id": "environment", "extension_id": "extension", "extension": map[string]any{"name": "Prior", "srcdoc": "<p>planned</p>", "field_types": []any{}, "sidebar": false}},
			body:  readEndpointBody("Extension", map[string]any{"extension": map[string]any{"name": "Returned", "srcdoc": "<p>returned</p>", "fieldTypes": []any{}, "sidebar": false}}),
		},
		{
			name: "space enablements update", typeName: "contentful_space_enablements", method: http.MethodPut,
			endpoint: "/spaces/space/enablements", schema: SpaceEnablementsResourceSchema(ctx),
			prior: map[string]any{"id": "space", "space_id": "space", "cross_space_links": false, "space_templates": false, "studio_experiences": false, "suggest_concepts": false},
			body:  readEndpointBody("SpaceEnablement", map[string]any{"crossSpaceLinks": map[string]any{"enabled": true}, "spaceTemplates": map[string]any{"enabled": true}}),
		},
		{
			name: "role update", typeName: "contentful_role", method: http.MethodPut,
			endpoint: "/spaces/space/roles/role", schema: RoleResourceSchema(ctx),
			prior: map[string]any{"id": "space/role", "space_id": "space", "role_id": "role", "name": "Prior", "permissions": map[string]any{"ContentModel": []any{"read"}}, "policies": []any{map[string]any{"actions": []any{"read"}, "effect": "allow"}}},
			body:  readEndpointBody("Role", map[string]any{"name": "Returned", "permissions": map[string]any{}}),
		},
		{
			name: "team update", typeName: "contentful_team", method: http.MethodPut,
			endpoint: "/organizations/organization/teams/team", schema: TeamResourceSchema(ctx),
			prior: map[string]any{"id": "organization/team", "organization_id": "organization", "team_id": "team", "name": "Team"},
			body:  readEndpointBody("Team", map[string]any{"name": "Returned", "description": ""}),
		},
		{
			name: "environment update", typeName: "contentful_environment", method: http.MethodPut,
			endpoint: "/spaces/space/environments/environment", schema: EnvironmentResourceSchema(ctx),
			prior: map[string]any{"id": "space/environment", "space_id": "space", "environment_id": "environment", "name": "Environment", "status": "ready"},
			body:  readEndpointBody("Environment", map[string]any{"name": "Returned"}),
		},
		{
			name: "environment alias update", typeName: "contentful_environment_alias", method: http.MethodPut,
			endpoint: "/spaces/space/environment_aliases/alias", schema: EnvironmentAliasResourceSchema(ctx),
			prior: map[string]any{"id": "space/alias", "space_id": "space", "environment_alias_id": "alias", "target_environment_id": "environment"},
			body:  readEndpointBody("EnvironmentAlias", map[string]any{"environment": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Environment", "id": "returned-target"}}}),
		},
		{
			name: "editor interface update", typeName: "contentful_editor_interface", method: http.MethodPut,
			endpoint: "/spaces/space/environments/environment/content_types/content-type/editor_interface", schema: EditorInterfaceResourceSchema(ctx),
			prior: map[string]any{"id": "space/environment/content-type", "space_id": "space", "environment_id": "environment", "content_type_id": "content-type"},
			body:  readEndpointBody("EditorInterface", map[string]any{"controls": []any{}}),
		},
		{
			name: "locale update", typeName: "contentful_locale", method: http.MethodPut,
			endpoint: "/spaces/space/environments/environment/locales/locale", schema: LocaleResourceSchema(ctx),
			prior: map[string]any{"id": "space/environment/locale", "space_id": "space", "environment_id": "environment", "locale_id": "locale", "name": "English (Australia)", "code": "en-AU", "content_delivery_api": true, "content_management_api": true, "optional": false, "default": false},
			body:  readEndpointBody("Locale", map[string]any{"name": "Returned", "code": "en-US", "fallbackCode": nil, "default": false, "contentDeliveryApi": true, "contentManagementApi": true, "optional": true}),
		},
		{
			name: "tag update", typeName: "contentful_tag", method: http.MethodPut,
			endpoint: "/spaces/space/environments/environment/tags/tag", schema: TagResourceSchema(ctx),
			prior: map[string]any{"id": "space/environment/tag", "space_id": "space", "environment_id": "environment", "tag_id": "tag", "name": "Tag", "visibility": "private"},
			body:  readEndpointBody("Tag", map[string]any{"name": "Returned"}),
		},
		{
			name: "taxonomy concept update", typeName: "contentful_taxonomy_concept", method: http.MethodPatch,
			endpoint: "/organizations/organization-id/taxonomy/concepts/concept", schema: TaxonomyConceptResourceSchema(ctx),
			prior:   map[string]any{"id": "organization-id/concept", "organization_id": "organization-id", "concept_id": "concept", "pref_label": map[string]any{"en-US": "Prior"}},
			changes: map[string]any{"pref_label": map[string]any{"en-US": "Planned"}},
			body:    readEndpointBody("TaxonomyConcept", map[string]any{"prefLabel": map[string]any{"en-US": "Returned"}, "conceptSchemes": []any{}}),
		},
		{
			name: "taxonomy concept scheme update", typeName: "contentful_taxonomy_concept_scheme", method: http.MethodPatch,
			endpoint: "/organizations/organization-id/taxonomy/concept-schemes/scheme", schema: TaxonomyConceptSchemeResourceSchema(ctx),
			prior:   map[string]any{"id": "organization-id/scheme", "organization_id": "organization-id", "concept_scheme_id": "scheme", "pref_label": map[string]any{"en-US": "Prior"}},
			changes: map[string]any{"pref_label": map[string]any{"en-US": "Planned"}},
			body:    readEndpointBody("TaxonomyConceptScheme", map[string]any{"prefLabel": map[string]any{"en-US": "Returned"}, "topConcepts": []any{}, "concepts": []any{}, "totalConcepts": 0}),
		},
	}
}

func TestPendingLifecycleMutationsRevokeAuthorityOnIdentityContradiction(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		endpoint, pendingKey, diagnosticSummary string
		diagnosticAttribute                     *tftypes.AttributePath
	}{
		"contentful_entry": {
			endpoint: "/spaces/space/environments/environment/entries/entry", pendingKey: "pending_publication_version",
			diagnosticSummary: "Unexpected entry identity",
		},
		"contentful_content_type": {
			endpoint: "/spaces/space/environments/environment/content_types/content-type", pendingKey: "pending_activation_version",
			diagnosticSummary: "Unexpected Contentful content type response", diagnosticAttribute: tftypes.NewAttributePath().WithAttributeName("space_id"),
		},
	}

	for _, fixture := range versionedMutationIdentityCases(t) {
		test, ok := cases[fixture.typeName]
		if !ok {
			continue
		}

		for _, stage := range []struct {
			name    string
			publish bool
		}{
			{name: "draft"},
			{name: "publication", publish: true},
		} {
			t.Run(fixture.typeName+"/"+stage.name, func(t *testing.T) {
				t.Parallel()

				plannedModel := maps.Clone(fixture.prior)
				maps.Copy(plannedModel, fixture.changes)
				retryModel := fixture.prior
				endpoint := test.endpoint

				if stage.publish {
					// An unchanged Plan consumes the existing marker without a draft PUT.
					plannedModel, retryModel = fixture.prior, plannedModel
					endpoint += "/published"
				}

				body, err := json.Marshal(pendingLifecycleIdentityResponse(fixture.typeName, stage.publish))
				require.NoError(t, err)

				var requests atomic.Int64

				protocol := configuredProtocolServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests.Add(1)
					assert.Equal(t, http.MethodPut, r.Method)
					assert.Equal(t, endpoint, r.URL.Path)
					assert.Equal(t, "31", r.Header.Get("X-Contentful-Version"))
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write(body)
				}))
				prior := mutationEndpointDynamicValue(t, fixture.schema, fixture.prior)
				plan := mutationEndpointDynamicValue(t, fixture.schema, plannedModel)
				private, err := json.Marshal(map[string][]byte{"version": []byte("31"), test.pendingKey: []byte("31")})
				require.NoError(t, err)

				response, err := protocol.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{
					TypeName: fixture.typeName, PriorState: &prior, PlannedState: &plan, Config: &plan, PlannedPrivate: private,
				})
				require.NoError(t, err)
				require.EqualValues(t, 1, requests.Load(), response.Diagnostics)
				require.Len(t, response.Diagnostics, 1, "only the space identity differs")
				assert.Equal(t, tfprotov6.DiagnosticSeverityError, response.Diagnostics[0].Severity)
				assert.Equal(t, test.diagnosticSummary, response.Diagnostics[0].Summary)
				assert.Equal(t, test.diagnosticAttribute, response.Diagnostics[0].Attribute)

				var saved map[string][]byte
				if len(response.Private) > 0 {
					require.NoError(t, json.Unmarshal(response.Private, &saved))
				}

				assert.Empty(t, saved["version"], "neither the prior nor returned version may authorize another write")
				assert.Empty(t, saved[test.pendingKey], "the existing marker must be revoked")
				require.NotNil(t, response.NewState)
				require.NotNil(t, response.NewIdentity)

				// A different Plan requires another write, rather than a state-only no-op.
				retryPlan := mutationEndpointDynamicValue(t, fixture.schema, retryModel)
				retry, err := protocol.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{
					TypeName: fixture.typeName, PriorState: response.NewState, PlannedIdentity: response.NewIdentity,
					PlannedState: &retryPlan, Config: &retryPlan, PlannedPrivate: response.Private,
				})
				require.NoError(t, err)
				assert.EqualValues(t, 1, requests.Load(), "revoked authority must stop retry before HTTP")
				require.Len(t, retry.Diagnostics, 1)
				assert.Equal(t, "Private version is unavailable", retry.Diagnostics[0].Summary)
			})
		}
	}
}

// Independently authored values match the Plan except for space identity. Both
// draft and publication tuples are valid, so unrelated errors cannot revoke authority.
func pendingLifecycleIdentityResponse(typeName string, publish bool) map[string]any {
	sys := map[string]any{
		"version":     73,
		"space":       map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Space", "id": "other-space"}},
		"environment": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Environment", "id": "environment"}},
	}
	body := map[string]any{"sys": sys}

	if typeName == "contentful_entry" {
		sys["type"], sys["id"] = "Entry", "entry"
		sys["contentType"] = map[string]any{"sys": map[string]any{"type": "Link", "linkType": "ContentType", "id": "content-type"}}

		body["fields"] = map[string]any{}
		if !publish {
			body["fields"] = map[string]any{"managed": map[string]any{"en-US": "changed"}}
		}

		body["metadata"] = map[string]any{"tags": []any{}, "concepts": []any{}}
	} else {
		sys["type"], sys["id"], sys["publishedVersion"] = "ContentType", "content-type", 1

		body["name"], body["description"], body["displayField"] = "Prior", "A test content type", "title"
		if !publish {
			body["name"] = "Planned"
		}

		body["fields"] = []any{map[string]any{
			"id": "title", "name": "Title", "type": "Symbol", "required": true,
			"localized": false, "disabled": false, "omitted": false, "validations": []any{},
		}}
	}

	if publish {
		sys["publishedVersion"] = 31
	}

	return body
}
