package provider_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type privateVersionState struct {
	name            string
	value           []byte
	expectedSummary string
}

type privateVersionRequest struct {
	method, path, version string
	versionPresent        bool
}

type privateVersionForwardingCase struct {
	requiredPrivateVersionResourceCase

	method string
	path   string
}

func TestPrivateVersionValuesAreForwardedToRepresentativeMutations(t *testing.T) {
	t.Parallel()

	selected := map[string]struct {
		method string
		path   string
	}{
		"delivery API key update": {method: http.MethodPut, path: "/spaces/space/api_keys/key"},
		"locale update":           {method: http.MethodPut, path: "/spaces/space/environments/environment/locales/locale"},
		"tag update":              {method: http.MethodPut, path: "/spaces/space/environments/environment/tags/tag"},
	}

	var cases []privateVersionForwardingCase

	for _, test := range requiredPrivateVersionResourceCases(t) {
		expected, ok := selected[test.name]
		if ok {
			cases = append(cases, privateVersionForwardingCase{
				requiredPrivateVersionResourceCase: test, method: expected.method, path: expected.path,
			})
			delete(selected, test.name)
		}
	}

	require.Empty(t, selected, "missing representative private version resource cases")

	cases = append(cases,
		privateVersionForwardingCase{
			requiredPrivateVersionResourceCase: requiredPrivateVersionResourceCase{
				name: "taxonomy concept delete", typeName: "contentful_taxonomy_concept",
				resourceSchema: TaxonomyConceptResourceSchema(t.Context()), model: taxonomyConceptDeleteState(), delete: true,
			},
			method: http.MethodDelete, path: "/organizations/organization-id/taxonomy/concepts/concept",
		},
		privateVersionForwardingCase{
			requiredPrivateVersionResourceCase: requiredPrivateVersionResourceCase{
				name: "taxonomy concept scheme delete", typeName: "contentful_taxonomy_concept_scheme",
				resourceSchema: TaxonomyConceptSchemeResourceSchema(t.Context()), model: taxonomyConceptSchemeDeleteState(), delete: true,
			},
			method: http.MethodDelete, path: "/organizations/organization-id/taxonomy/concept-schemes/scheme",
		},
	)

	for _, test := range cases {
		for _, version := range []int{0, -1} {
			t.Run(test.name+"/version "+strconv.Itoa(version), func(t *testing.T) {
				t.Parallel()

				var requests []privateVersionRequest

				testServer := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
					_, versionPresent := request.Header["X-Contentful-Version"]
					requests = append(requests, privateVersionRequest{
						method: request.Method, path: request.URL.Path,
						version: request.Header.Get("X-Contentful-Version"), versionPresent: versionPresent,
					})

					response.WriteHeader(http.StatusNoContent)
				}))
				t.Cleanup(testServer.Close)

				providerServer, err := makeTestAccProtoV6ProviderFactories(
					WithContentfulURL(testServer.URL),
					WithAccessToken("test-token"),
				)["contentful"]()
				require.NoError(t, err)

				providerConfig, err := providerConfigDynamicValue(map[string]any{
					"url":          tftypes.UnknownValue,
					"access_token": tftypes.UnknownValue,
				})
				require.NoError(t, err)

				configureResponse, err := providerServer.ConfigureProvider(t.Context(), &tfprotov6.ConfigureProviderRequest{
					Config: &providerConfig,
				})
				require.NoError(t, err)
				require.Empty(t, configureResponse.Diagnostics)

				state := resourceModelDynamicValue(t, test.resourceSchema, test.model)

				plannedState := state
				if test.plannedModel != nil {
					plannedState = resourceModelDynamicValue(t, test.resourceSchema, test.plannedModel)
				}

				if test.delete {
					plannedState = nullResourceDynamicValue(t, test.resourceSchema)
				}

				_, err = providerServer.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{
					TypeName: test.typeName, PriorState: &state, PlannedState: &plannedState,
					Config: &plannedState, PlannedPrivate: privateVersionBytes(t, version),
				})
				require.NoError(t, err)
				require.Equal(t, []privateVersionRequest{{
					method: test.method, path: test.path, version: strconv.Itoa(version), versionPresent: true,
				}}, requests)
			})
		}
	}
}

//nolint:tparallel // Resource subtests intentionally share a serial HTTP request journal.
func TestRequiredPrivateVersionErrorsStopBeforeMutation(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int64

	testServer := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		requestCount.Add(1)
		http.Error(response, "unexpected Contentful request", http.StatusInternalServerError)
	}))
	t.Cleanup(testServer.Close)

	providerServer, err := makeTestAccProtoV6ProviderFactories(
		WithContentfulURL(testServer.URL),
		WithAccessToken("test-token"),
	)["contentful"]()
	require.NoError(t, err)

	providerConfig, err := providerConfigDynamicValue(map[string]any{
		"url":          tftypes.UnknownValue,
		"access_token": tftypes.UnknownValue,
	})
	require.NoError(t, err)

	configureResponse, err := providerServer.ConfigureProvider(t.Context(), &tfprotov6.ConfigureProviderRequest{
		Config: &providerConfig,
	})
	require.NoError(t, err)
	require.Empty(t, configureResponse.Diagnostics)

	malformedPrivate, err := json.Marshal(map[string][]byte{"version": []byte(`"invalid"`)})
	require.NoError(t, err)

	// Subtests share one request journal so they deliberately remain serial.
	//nolint:paralleltest
	for _, test := range requiredPrivateVersionResourceCases(t) {
		t.Run(test.name, func(t *testing.T) {
			state := resourceModelDynamicValue(t, test.resourceSchema, test.model)

			plannedState := state
			if test.plannedModel != nil {
				plannedState = resourceModelDynamicValue(t, test.resourceSchema, test.plannedModel)
			}

			if test.delete {
				plannedState = nullResourceDynamicValue(t, test.resourceSchema)
			}

			privateStates := []privateVersionState{
				{name: "missing", value: []byte(`{}`), expectedSummary: "Private version is unavailable"},
				{name: "malformed", value: malformedPrivate, expectedSummary: "Failed to unmarshal value"},
			}

			for _, privateState := range privateStates {
				t.Run(privateState.name, func(t *testing.T) {
					requestCount.Store(0)

					response, err := providerServer.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{
						TypeName:       test.typeName,
						PriorState:     &state,
						PlannedState:   &plannedState,
						Config:         &plannedState,
						PlannedPrivate: privateState.value,
					})
					require.NoError(t, err)
					require.NotEmpty(t, response.Diagnostics)
					assert.Equal(t, tfprotov6.DiagnosticSeverityError, response.Diagnostics[0].Severity)

					diagnosticSummaries := make([]string, 0, len(response.Diagnostics))
					for _, diagnostic := range response.Diagnostics {
						diagnosticSummaries = append(diagnosticSummaries, diagnostic.Summary)
					}

					assert.Contains(t, diagnosticSummaries, privateState.expectedSummary)
					assert.Zero(t, requestCount.Load())
				})
			}
		})
	}
}

type requiredPrivateVersionResourceCase struct {
	name           string
	typeName       string
	resourceSchema schema.Schema
	model          any
	plannedModel   any
	delete         bool
}

func requiredPrivateVersionResourceCases(t *testing.T) []requiredPrivateVersionResourceCase {
	t.Helper()

	webhook := validWebhookRequestModel()
	webhook.ID = types.StringValue("space/webhook")
	webhook.SpaceID = types.StringValue("space")
	webhook.WebhookID = types.StringValue("webhook")
	webhook.Timeouts = TimeoutsNull()

	extension := validExtensionRequestModel()
	extension.ID = types.StringValue("space/environment/extension")
	extension.SpaceID = types.StringValue("space")
	extension.EnvironmentID = types.StringValue("environment")
	extension.ExtensionID = types.StringValue("extension")
	extension.Parameters = jsontypes.NewNormalizedNull()
	extension.Timeouts = TimeoutsNull()

	role := validRoleRequestModel()
	role.ID = types.StringValue("space/role")
	role.SpaceID = types.StringValue("space")
	role.RoleID = types.StringValue("role")
	role.Description = types.StringNull()
	role.Timeouts = TimeoutsNull()

	editorInterface := newNullEditorInterfaceModel()
	editorInterface.ID = types.StringValue("space/environment/content-type")
	editorInterface.SpaceID = types.StringValue("space")
	editorInterface.EnvironmentID = types.StringValue("environment")
	editorInterface.ContentTypeID = types.StringValue("content-type")
	editorInterface.Timeouts = updateOnlyTimeoutsNull()

	tag := TagModel{
		IDIdentityModel: IDIdentityModel{ID: types.StringValue("space/environment/tag")},
		TagIdentityModel: TagIdentityModel{
			SpaceID:       types.StringValue("space"),
			EnvironmentID: types.StringValue("environment"),
			TagID:         types.StringValue("tag"),
		},
		Name:       types.StringValue("Tag"),
		Visibility: types.StringValue("private"),
		Timeouts:   TimeoutsNull(),
	}
	locale := LocaleModel{
		IDIdentityModel: IDIdentityModel{ID: types.StringValue("space/environment/locale")},
		LocaleIdentityModel: LocaleIdentityModel{
			SpaceID:       types.StringValue("space"),
			EnvironmentID: types.StringValue("environment"),
			LocaleID:      types.StringValue("locale"),
		},
		Name:                 types.StringValue("English (Australia)"),
		Code:                 types.StringValue("en-AU"),
		FallbackCode:         types.StringNull(),
		ContentDeliveryAPI:   types.BoolValue(true),
		ContentManagementAPI: types.BoolValue(true),
		Optional:             types.BoolValue(false),
		Default:              types.BoolValue(false),
		Timeouts:             TimeoutsNull(),
	}

	contentType := validContentTypeRequestModel()
	contentType.ID = types.StringValue("space/environment/content-type")
	contentType.SpaceID = types.StringValue("space")
	contentType.EnvironmentID = types.StringValue("environment")
	contentType.ContentTypeID = types.StringValue("content-type")
	contentType.PublishedVersion = types.Int64Value(1)
	contentType.Timeouts = TimeoutsNull()
	plannedContentType := contentType
	plannedContentType.Name = types.StringValue("Changed content type")
	entry := EntryModel{
		IDIdentityModel:    IDIdentityModel{ID: types.StringValue("space/environment/entry")},
		EntryIdentityModel: NewEntryIdentityModel("space", "environment", "entry"),
		ContentTypeID:      types.StringValue("content-type"),
		Fields:             NewTypedMap(map[string]jsontypes.Normalized{}),
		Metadata: NewTypedObject(EntryMetadataValue{
			Concepts: NewTypedList([]types.String{}),
			Tags:     NewTypedList([]types.String{}),
		}),
		Timeouts: TimeoutsNull(),
	}
	plannedEntry := entry
	plannedEntry.Fields = NewTypedMap(map[string]jsontypes.Normalized{
		"managed": jsontypes.NewNormalizedValue(`{"en-US":"changed"}`),
	})

	return []requiredPrivateVersionResourceCase{
		{
			name:           "content type update",
			typeName:       "contentful_content_type",
			resourceSchema: ContentTypeResourceSchema(t.Context()),
			model:          contentType,
			plannedModel:   plannedContentType,
		},
		{
			name:           "entry update",
			typeName:       "contentful_entry",
			resourceSchema: EntryResourceSchema(t.Context()),
			model:          entry,
			plannedModel:   plannedEntry,
		},
		{
			name:           "delivery API key update",
			typeName:       "contentful_delivery_api_key",
			resourceSchema: DeliveryAPIKeyResourceSchema(t.Context()),
			model: DeliveryAPIKeyModel{
				IDIdentityModel: IDIdentityModel{ID: types.StringValue("space/key")},
				DeliveryAPIKeyIdentityModel: DeliveryAPIKeyIdentityModel{
					SpaceID:  types.StringValue("space"),
					APIKeyID: types.StringValue("key"),
				},
				Name:            types.StringValue("Key"),
				Description:     types.StringNull(),
				Environments:    NewTypedList([]types.String{}),
				AccessToken:     types.StringValue("redacted"),
				PreviewAPIKeyID: types.StringValue("preview"),
				Timeouts:        TimeoutsNull(),
			},
		},
		{name: "webhook update", typeName: "contentful_webhook", resourceSchema: WebhookResourceSchema(t.Context()), model: webhook},
		{name: "extension update", typeName: "contentful_extension", resourceSchema: ExtensionResourceSchema(t.Context()), model: extension},
		{
			name:           "space enablements update",
			typeName:       "contentful_space_enablements",
			resourceSchema: SpaceEnablementsResourceSchema(t.Context()),
			model: SpaceEnablementsModel{
				IDIdentityModel:               IDIdentityModel{ID: types.StringValue("space")},
				SpaceEnablementsIdentityModel: SpaceEnablementsIdentityModel{SpaceID: types.StringValue("space")},
				CrossSpaceLinks:               types.BoolValue(false),
				SpaceTemplates:                types.BoolValue(false),
				StudioExperiences:             types.BoolValue(false),
				SuggestConcepts:               types.BoolValue(false),
				Timeouts:                      updateOnlyTimeoutsNull(),
			},
		},
		{name: "role update", typeName: "contentful_role", resourceSchema: RoleResourceSchema(t.Context()), model: role},
		{
			name:           "team update",
			typeName:       "contentful_team",
			resourceSchema: TeamResourceSchema(t.Context()),
			model: TeamModel{
				IDIdentityModel:   IDIdentityModel{ID: types.StringValue("organization/team")},
				TeamIdentityModel: TeamIdentityModel{OrganizationID: types.StringValue("organization"), TeamID: types.StringValue("team")},
				Name:              types.StringValue("Team"),
				Description:       types.StringNull(),
				Timeouts:          TimeoutsNull(),
			},
		},
		{
			name:           "environment update",
			typeName:       "contentful_environment",
			resourceSchema: EnvironmentResourceSchema(t.Context()),
			model: EnvironmentModel{
				IDIdentityModel:          IDIdentityModel{ID: types.StringValue("space/environment")},
				EnvironmentIdentityModel: EnvironmentIdentityModel{SpaceID: types.StringValue("space"), EnvironmentID: types.StringValue("environment")},
				Name:                     types.StringValue("Environment"),
				Status:                   types.StringValue("ready"),
				SourceEnvironmentID:      types.StringNull(),
				Timeouts:                 TimeoutsNull(),
			},
		},
		{
			name:           "environment alias update",
			typeName:       "contentful_environment_alias",
			resourceSchema: EnvironmentAliasResourceSchema(t.Context()),
			model: EnvironmentAliasModel{
				IDIdentityModel:               IDIdentityModel{ID: types.StringValue("space/alias")},
				EnvironmentAliasIdentityModel: EnvironmentAliasIdentityModel{SpaceID: types.StringValue("space"), EnvironmentAliasID: types.StringValue("alias")},
				TargetEnvironmentID:           types.StringValue("environment"),
				Timeouts:                      TimeoutsNull(),
			},
		},
		{name: "editor interface update", typeName: "contentful_editor_interface", resourceSchema: EditorInterfaceResourceSchema(t.Context()), model: editorInterface},
		{name: "locale update", typeName: "contentful_locale", resourceSchema: LocaleResourceSchema(t.Context()), model: locale},
		{name: "tag update", typeName: "contentful_tag", resourceSchema: TagResourceSchema(t.Context()), model: tag},
		{name: "tag delete", typeName: "contentful_tag", resourceSchema: TagResourceSchema(t.Context()), model: tag, delete: true},
	}
}

func resourceModelDynamicValue(t *testing.T, resourceSchema schema.Schema, model any) tfprotov6.DynamicValue {
	t.Helper()

	plan := tfsdk.Plan{Schema: resourceSchema}
	diagnostics := plan.Set(t.Context(), model)
	require.False(t, diagnostics.HasError(), diagnostics.Errors())

	value, err := tfprotov6.NewDynamicValue(resourceSchema.Type().TerraformType(t.Context()), plan.Raw)
	require.NoError(t, err)

	return value
}

func nullResourceDynamicValue(t *testing.T, resourceSchema schema.Schema) tfprotov6.DynamicValue {
	t.Helper()

	terraformType := resourceSchema.Type().TerraformType(t.Context())
	value, err := tfprotov6.NewDynamicValue(terraformType, tftypes.NewValue(terraformType, nil))
	require.NoError(t, err)

	return value
}

func privateVersionBytes(t *testing.T, version int) []byte {
	t.Helper()

	encodedVersion, err := json.Marshal(version)
	require.NoError(t, err)
	private, err := json.Marshal(map[string][]byte{"version": encodedVersion})
	require.NoError(t, err)

	return private
}

func taxonomyConceptDeleteState() TaxonomyConceptModel {
	listType := types.ListType{ElemType: types.StringType}

	return TaxonomyConceptModel{
		IDIdentityModel:              NewIDIdentityModelFromMultipartID("organization-id", "concept"),
		TaxonomyConceptIdentityModel: TaxonomyConceptIdentityModel{OrganizationID: types.StringValue("organization-id"), ConceptID: types.StringValue("concept")},
		URI:                          types.StringNull(),
		PrefLabel:                    types.MapNull(types.StringType),
		AltLabels:                    types.MapNull(listType),
		HiddenLabels:                 types.MapNull(listType),
		Notations:                    types.ListNull(types.StringType),
		Note:                         types.MapNull(types.StringType),
		ChangeNote:                   types.MapNull(types.StringType),
		Definition:                   types.MapNull(types.StringType),
		EditorialNote:                types.MapNull(types.StringType),
		Example:                      types.MapNull(types.StringType),
		HistoryNote:                  types.MapNull(types.StringType),
		ScopeNote:                    types.MapNull(types.StringType),
		BroaderConceptIDs:            types.ListNull(types.StringType),
		RelatedConceptIDs:            types.ListNull(types.StringType),
		ConceptSchemeIDs:             types.SetNull(types.StringType),
		Timeouts:                     TimeoutsNull(),
	}
}

func taxonomyConceptSchemeDeleteState() TaxonomyConceptSchemeModel {
	return TaxonomyConceptSchemeModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID("organization-id", "scheme"),
		TaxonomyConceptSchemeIdentityModel: TaxonomyConceptSchemeIdentityModel{
			OrganizationID: types.StringValue("organization-id"), ConceptSchemeID: types.StringValue("scheme"),
		},
		URI:           types.StringNull(),
		PrefLabel:     types.MapNull(types.StringType),
		Definition:    types.MapNull(types.StringType),
		TopConceptIDs: types.ListNull(types.StringType),
		ConceptIDs:    types.ListNull(types.StringType),
		TotalConcepts: types.Int64Null(),
		Timeouts:      TimeoutsNull(),
	}
}

func updateOnlyTimeoutsNull() timeouts.Value {
	return timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{
		"create": types.StringType,
		"read":   types.StringType,
		"update": types.StringType,
	})}
}
