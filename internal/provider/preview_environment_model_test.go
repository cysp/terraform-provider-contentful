package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func previewEnvironmentModel(configurations map[string]string) PreviewEnvironmentModel {
	values := make(map[string]TypedObject[PreviewEnvironmentContentTypeConfigurationValue], len(configurations))
	for contentTypeID, url := range configurations {
		values[contentTypeID] = NewTypedObject(PreviewEnvironmentContentTypeConfigurationValue{
			URL: types.StringValue(url),
		})
	}

	return PreviewEnvironmentModel{
		Name:                      types.StringValue("Preview"),
		Description:               types.StringValue(""),
		ContentTypeConfigurations: NewTypedMap(values),
	}
}

func TestPreviewEnvironmentModelFiltersDisabledConfigurations(t *testing.T) {
	t.Parallel()

	response := cm.PreviewEnvironment{
		Sys:         cm.NewPreviewEnvironmentSys("space", "preview"),
		Name:        "Preview",
		Description: "",
		Configurations: []cm.PreviewEnvironmentConfiguration{
			{
				URL:         "https://preview.invalid/page",
				ContentType: cm.NewOptString("page"),
				EntityType:  cm.NewOptString("ContentType"),
				EntityId:    cm.NewOptString("page"),
				Enabled:     true,
			},
			{
				URL:         "https://preview.invalid/author",
				ContentType: cm.NewOptString("author"),
				EntityType:  cm.NewOptString("ContentType"),
				EntityId:    cm.NewOptString("author"),
				Enabled:     false,
			},
		},
	}

	model, diagnostics := NewPreviewEnvironmentModelFromResponse(t.Context(), response)
	require.False(t, diagnostics.HasError())
	assert.Equal(t, "space/preview", model.ID.ValueString())
	require.Len(t, model.ContentTypeConfigurations.Elements(), 1)
	assert.Equal(
		t,
		"https://preview.invalid/page",
		model.ContentTypeConfigurations.Elements()["page"].Value().URL.ValueString(),
	)

	request, diagnostics := model.ToPreviewEnvironmentData(t.Context(), path.Empty())
	require.False(t, diagnostics.HasError())
	require.Len(t, request.Configurations, 1)
	assert.Equal(t, "page", request.Configurations[0].EntityId)
	assert.Equal(t, "ContentType", request.Configurations[0].EntityType)
	assert.True(t, request.Configurations[0].Enabled)
	assert.False(t, request.Configurations[0].Example.IsSet())
}

func TestPreviewEnvironmentModelResponseIdentityNormalization(t *testing.T) {
	t.Parallel()

	tests := map[string]cm.PreviewEnvironmentConfiguration{
		"legacy content type": {
			URL:         "https://preview.invalid/page",
			ContentType: cm.NewOptString("page"),
			Enabled:     true,
		},
		"entity ID": {
			URL:      "https://preview.invalid/page",
			EntityId: cm.NewOptString("page"),
			Enabled:  true,
		},
		"matching aliases": {
			URL:         "https://preview.invalid/page",
			EntityType:  cm.NewOptString("ContentType"),
			EntityId:    cm.NewOptString("page"),
			ContentType: cm.NewOptString("page"),
			Enabled:     true,
		},
	}

	for name, configuration := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model, diagnostics := NewPreviewEnvironmentModelFromResponse(t.Context(), cm.PreviewEnvironment{
				Sys:            cm.NewPreviewEnvironmentSys("space", "preview"),
				Name:           "Preview",
				Description:    "",
				Configurations: []cm.PreviewEnvironmentConfiguration{configuration},
			})
			require.False(t, diagnostics.HasError())
			assert.Equal(
				t,
				types.StringValue("https://preview.invalid/page"),
				model.ContentTypeConfigurations.Elements()["page"].Value().URL,
			)
		})
	}
}

func TestPreviewEnvironmentModelResponseWarnsAndPreservesRepresentableSiblings(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		configuration cm.PreviewEnvironmentConfiguration
		expectedPath  string
	}{
		"missing identity": {
			configuration: cm.PreviewEnvironmentConfiguration{URL: "https://preview.invalid/invalid", Enabled: true},
			expectedPath:  "content_type_configurations",
		},
		"conflicting aliases": {
			configuration: cm.PreviewEnvironmentConfiguration{
				URL:         "https://preview.invalid/page",
				EntityType:  cm.NewOptString("ContentType"),
				EntityId:    cm.NewOptString("page"),
				ContentType: cm.NewOptString("article"),
				Enabled:     true,
			},
			expectedPath: `content_type_configurations["page"]`,
		},
		"unsupported enabled entity": {
			configuration: cm.PreviewEnvironmentConfiguration{
				URL:        "https://preview.invalid/page",
				EntityType: cm.NewOptString("Entry"),
				EntityId:   cm.NewOptString("page"),
				Enabled:    true,
			},
			expectedPath: `content_type_configurations["page"]`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model, diagnostics := NewPreviewEnvironmentModelFromResponse(t.Context(), cm.PreviewEnvironment{
				Sys:         cm.NewPreviewEnvironmentSys("space", "preview"),
				Name:        "Preview",
				Description: "",
				Configurations: []cm.PreviewEnvironmentConfiguration{
					{
						URL:        "https://preview.invalid/author",
						EntityType: cm.NewOptString("ContentType"),
						EntityId:   cm.NewOptString("author"),
						Enabled:    true,
					},
					test.configuration,
				},
			})
			require.False(t, diagnostics.HasError())
			require.Len(t, diagnostics.Warnings(), 1)
			diagnostic, ok := diagnostics.Warnings()[0].(diag.DiagnosticWithPath)
			require.True(t, ok)
			assert.Equal(t, test.expectedPath, diagnostic.Path().String())
			require.Len(t, model.ContentTypeConfigurations.Elements(), 1)
			assert.Equal(
				t,
				types.StringValue("https://preview.invalid/author"),
				model.ContentTypeConfigurations.Elements()["author"].Value().URL,
			)
		})
	}
}

func TestPreviewEnvironmentModelResponseKeepsFirstActiveDuplicate(t *testing.T) {
	t.Parallel()

	model, diagnostics := NewPreviewEnvironmentModelFromResponse(t.Context(), cm.PreviewEnvironment{
		Sys:         cm.NewPreviewEnvironmentSys("space", "preview"),
		Name:        "Preview",
		Description: "",
		Configurations: []cm.PreviewEnvironmentConfiguration{
			previewEnvironmentResponseConfiguration("page", "https://preview.invalid/first", true),
			previewEnvironmentResponseConfiguration("page", "https://preview.invalid/second", true),
			{
				URL:        "https://preview.invalid/disabled",
				EntityType: cm.NewOptString("Entry"),
				EntityId:   cm.NewOptString("unsupported-disabled"),
				Enabled:    false,
			},
		},
	})

	require.False(t, diagnostics.HasError())
	require.Len(t, diagnostics.Warnings(), 1)
	diagnostic, ok := diagnostics.Warnings()[0].(diag.DiagnosticWithPath)
	require.True(t, ok)
	assert.Equal(t, `content_type_configurations["page"]`, diagnostic.Path().String())
	require.Len(t, model.ContentTypeConfigurations.Elements(), 1)
	assert.Equal(
		t,
		types.StringValue("https://preview.invalid/first"),
		model.ContentTypeConfigurations.Elements()["page"].Value().URL,
	)
}

func TestPreviewEnvironmentModelResponseRequiresResourceIdentity(t *testing.T) {
	t.Parallel()

	for name, response := range map[string]cm.PreviewEnvironment{
		"missing space ID": {
			Sys: cm.NewPreviewEnvironmentSys("", "preview"),
		},
		"missing preview environment ID": {
			Sys: cm.NewPreviewEnvironmentSys("space", ""),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, diagnostics := NewPreviewEnvironmentModelFromResponse(t.Context(), response)
			require.True(t, diagnostics.HasError())
			require.Len(t, diagnostics.Errors(), 1)
			_, ok := diagnostics.Errors()[0].(diag.DiagnosticWithPath)
			require.True(t, ok)
		})
	}
}

func TestPreviewEnvironmentModelRequestRejectsUnknownNestedValues(t *testing.T) {
	t.Parallel()

	model := PreviewEnvironmentModel{
		Name:        types.StringValue("Preview"),
		Description: types.StringValue(""),
		ContentTypeConfigurations: NewTypedMap(map[string]TypedObject[PreviewEnvironmentContentTypeConfigurationValue]{
			"page": NewTypedObject(PreviewEnvironmentContentTypeConfigurationValue{
				URL: types.StringUnknown(),
			}),
		}),
	}

	_, diagnostics := model.ToPreviewEnvironmentData(t.Context(), path.Empty())
	require.True(t, diagnostics.HasError())
}

func TestPreviewEnvironmentModelRequestRejectsNullConfiguration(t *testing.T) {
	t.Parallel()

	model := PreviewEnvironmentModel{
		Name:        types.StringValue("Preview"),
		Description: types.StringValue(""),
		ContentTypeConfigurations: NewTypedMap(map[string]TypedObject[PreviewEnvironmentContentTypeConfigurationValue]{
			"page": NewTypedObjectNull[PreviewEnvironmentContentTypeConfigurationValue](),
		}),
	}

	_, diagnostics := model.ToPreviewEnvironmentData(t.Context(), path.Empty())
	require.True(t, diagnostics.HasError())
}

func TestPreviewEnvironmentModelRequestRejectsUnresolvedTopLevelValues(t *testing.T) {
	t.Parallel()

	tests := map[string]func(*PreviewEnvironmentModel){
		"unknown name": func(model *PreviewEnvironmentModel) {
			model.Name = types.StringUnknown()
		},
		"null name": func(model *PreviewEnvironmentModel) {
			model.Name = types.StringNull()
		},
		"unknown description": func(model *PreviewEnvironmentModel) {
			model.Description = types.StringUnknown()
		},
		"null description": func(model *PreviewEnvironmentModel) {
			model.Description = types.StringNull()
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := previewEnvironmentModel(map[string]string{
				"page": "https://preview.invalid/page",
			})
			mutate(&model)

			request, diagnostics := model.ToPreviewEnvironmentData(t.Context(), path.Empty())
			require.True(t, diagnostics.HasError())
			assert.Empty(t, request.Name)
			assert.Empty(t, request.Description)
			assert.Empty(t, request.Configurations)

			state := previewEnvironmentModel(map[string]string{
				"page": "https://preview.invalid/page",
			})
			plan := previewEnvironmentModel(map[string]string{
				"page": "https://preview.invalid/page-new",
			})
			mutate(&plan)

			request, diagnostics = ToPreviewEnvironmentUpdateData(
				t.Context(),
				path.Empty(),
				&state,
				&plan,
			)
			require.True(t, diagnostics.HasError())
			assert.Empty(t, request.Name)
			assert.Empty(t, request.Description)
			assert.Empty(t, request.Configurations)
		})
	}
}

func TestPreviewEnvironmentUpdateDataUsesStateToPlanDelta(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		state    map[string]string
		plan     map[string]string
		expected []cm.PreviewEnvironmentConfigurationData
	}{
		"unchanged": {
			state:    map[string]string{"page": "https://preview.invalid/page"},
			plan:     map[string]string{"page": "https://preview.invalid/page"},
			expected: []cm.PreviewEnvironmentConfigurationData{},
		},
		"add": {
			state: map[string]string{"page": "https://preview.invalid/page"},
			plan: map[string]string{
				"page":   "https://preview.invalid/page",
				"author": "https://preview.invalid/author",
			},
			expected: []cm.PreviewEnvironmentConfigurationData{
				previewEnvironmentConfigurationData("author", "https://preview.invalid/author", true),
			},
		},
		"change URL": {
			state: map[string]string{"page": "https://preview.invalid/page"},
			plan:  map[string]string{"page": "https://preview.invalid/page-new"},
			expected: []cm.PreviewEnvironmentConfigurationData{
				previewEnvironmentConfigurationData("page", "https://preview.invalid/page-new", true),
			},
		},
		"remove": {
			state: map[string]string{"page": "https://preview.invalid/page"},
			plan:  map[string]string{},
			expected: []cm.PreviewEnvironmentConfigurationData{
				previewEnvironmentConfigurationData("page", "https://preview.invalid/page", false),
			},
		},
		"mixed sorted delta": {
			state: map[string]string{
				"page":   "https://preview.invalid/page",
				"author": "https://preview.invalid/author",
			},
			plan: map[string]string{
				"author":   "https://preview.invalid/author-new",
				"category": "https://preview.invalid/category",
			},
			expected: []cm.PreviewEnvironmentConfigurationData{
				previewEnvironmentConfigurationData("author", "https://preview.invalid/author-new", true),
				previewEnvironmentConfigurationData("category", "https://preview.invalid/category", true),
				previewEnvironmentConfigurationData("page", "https://preview.invalid/page", false),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			state := previewEnvironmentModel(test.state)
			plan := previewEnvironmentModel(test.plan)
			request, diagnostics := ToPreviewEnvironmentUpdateData(t.Context(), path.Empty(), &state, &plan)
			require.False(t, diagnostics.HasError())
			assert.Equal(t, test.expected, request.Configurations)
		})
	}
}

func TestReconcilePreviewEnvironmentMutationResponseRestoresEquivalentPlan(t *testing.T) {
	t.Parallel()

	plan := previewEnvironmentMutationModel("preview", "planned description", map[string]string{
		"author": "https://preview.invalid/author",
		"page":   "https://preview.invalid/page",
	})
	response := previewEnvironmentResponse("space", "preview", "Planned", "planned description", map[string]string{
		"page":   "https://preview.invalid/page",
		"author": "https://preview.invalid/author",
	})

	state, responseDiagnostics, consistencyDiagnostics := ReconcilePreviewEnvironmentMutationResponse(
		t.Context(),
		response,
		plan,
		plan.PreviewEnvironmentIdentityModel,
	)

	assert.Empty(t, responseDiagnostics)
	assert.Empty(t, consistencyDiagnostics)
	assert.Equal(t, plan.Name, state.Name)
	assert.Equal(t, plan.Description, state.Description)
	assert.True(t, plan.ContentTypeConfigurations.Equal(state.ContentTypeConfigurations))
}

func TestReconcilePreviewEnvironmentMutationResponseDetectsEveryOwnedContradiction(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		response               cm.PreviewEnvironment
		expectedPath           string
		expectedName           string
		expectedDescription    string
		expectedConfigurations map[string]string
	}{
		"name": {
			response: previewEnvironmentResponse("space", "preview", "Remote", "Planned description", map[string]string{
				"page": "https://preview.invalid/page",
			}),
			expectedPath:           "name",
			expectedName:           "Remote",
			expectedDescription:    "Planned description",
			expectedConfigurations: map[string]string{"page": "https://preview.invalid/page"},
		},
		"description": {
			response: previewEnvironmentResponse("space", "preview", "Planned", "Remote description", map[string]string{
				"page": "https://preview.invalid/page",
			}),
			expectedPath:           "description",
			expectedName:           "Planned",
			expectedDescription:    "Remote description",
			expectedConfigurations: map[string]string{"page": "https://preview.invalid/page"},
		},
		"space identity": {
			response: previewEnvironmentResponse("other-space", "preview", "Planned", "Planned description", map[string]string{
				"page": "https://preview.invalid/page",
			}),
			expectedPath:           "space_id",
			expectedName:           "Planned",
			expectedDescription:    "Planned description",
			expectedConfigurations: map[string]string{"page": "https://preview.invalid/page"},
		},
		"selected identity": {
			response: previewEnvironmentResponse("space", "other-preview", "Planned", "Planned description", map[string]string{
				"page": "https://preview.invalid/page",
			}),
			expectedPath:           "preview_environment_id",
			expectedName:           "Planned",
			expectedDescription:    "Planned description",
			expectedConfigurations: map[string]string{"page": "https://preview.invalid/page"},
		},
		"changed or unchanged configuration contradicted": {
			response: previewEnvironmentResponse("space", "preview", "Planned", "Planned description", map[string]string{
				"page": "https://preview.invalid/remote-page",
			}),
			expectedPath:           `content_type_configurations["page"].url`,
			expectedName:           "Planned",
			expectedDescription:    "Planned description",
			expectedConfigurations: map[string]string{"page": "https://preview.invalid/remote-page"},
		},
		"planned configuration omitted": {
			response:               previewEnvironmentResponse("space", "preview", "Planned", "Planned description", map[string]string{}),
			expectedPath:           `content_type_configurations["page"]`,
			expectedName:           "Planned",
			expectedDescription:    "Planned description",
			expectedConfigurations: map[string]string{},
		},
		"removed or unexpected configuration remains active": {
			response: previewEnvironmentResponse("space", "preview", "Planned", "Planned description", map[string]string{
				"page":   "https://preview.invalid/page",
				"author": "https://preview.invalid/author",
			}),
			expectedPath:        `content_type_configurations["author"]`,
			expectedName:        "Planned",
			expectedDescription: "Planned description",
			expectedConfigurations: map[string]string{
				"page":   "https://preview.invalid/page",
				"author": "https://preview.invalid/author",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			plan := previewEnvironmentMutationModel("preview", "Planned description", map[string]string{
				"page": "https://preview.invalid/page",
			})
			state, responseDiagnostics, consistencyDiagnostics := ReconcilePreviewEnvironmentMutationResponse(
				t.Context(),
				test.response,
				plan,
				plan.PreviewEnvironmentIdentityModel,
			)

			assert.Empty(t, responseDiagnostics)
			require.True(t, consistencyDiagnostics.HasError())
			assert.Contains(t, attributeDiagnosticPaths(t, consistencyDiagnostics), test.expectedPath)

			assert.Equal(t, types.StringValue(test.expectedName), state.Name)
			assert.Equal(t, types.StringValue(test.expectedDescription), state.Description)
			expectedConfigurations := previewEnvironmentModel(test.expectedConfigurations).ContentTypeConfigurations
			assert.True(t, expectedConfigurations.Equal(state.ContentTypeConfigurations))
			assert.Equal(t, types.StringValue("space"), state.SpaceID)
			assert.Equal(t, types.StringValue("preview"), state.PreviewEnvironmentID)
			assert.Equal(t, types.StringValue("space/preview"), state.ID)
		})
	}
}

func TestReconcilePreviewEnvironmentMutationResponseHonorsSelectedIDOwnership(t *testing.T) {
	t.Parallel()

	plan := previewEnvironmentMutationModel("planned-id", "", map[string]string{})
	response := previewEnvironmentResponse("space", "generated-id", "Planned", "", map[string]string{})

	generatedState, responseDiagnostics, consistencyDiagnostics := ReconcilePreviewEnvironmentMutationResponse(
		t.Context(),
		response,
		plan,
		PreviewEnvironmentIdentityModel{SpaceID: plan.SpaceID, PreviewEnvironmentID: types.StringNull()},
	)
	assert.Empty(t, responseDiagnostics)
	assert.Empty(t, consistencyDiagnostics)
	assert.Equal(t, types.StringValue("generated-id"), generatedState.PreviewEnvironmentID)
	assert.Equal(t, types.StringValue("space/generated-id"), generatedState.ID)

	selectedState, responseDiagnostics, consistencyDiagnostics := ReconcilePreviewEnvironmentMutationResponse(
		t.Context(),
		response,
		plan,
		plan.PreviewEnvironmentIdentityModel,
	)
	assert.Empty(t, responseDiagnostics)
	require.True(t, consistencyDiagnostics.HasError())
	assert.Contains(t, attributeDiagnosticPaths(t, consistencyDiagnostics), "preview_environment_id")
	assert.Equal(t, types.StringValue("planned-id"), selectedState.PreviewEnvironmentID)
	assert.Equal(t, types.StringValue("space/planned-id"), selectedState.ID)
}

func TestReconcilePreviewEnvironmentMutationResponseRetainsProjectionWarnings(t *testing.T) {
	t.Parallel()

	plan := previewEnvironmentMutationModel("preview", "", map[string]string{
		"page": "https://preview.invalid/page",
	})
	response := previewEnvironmentResponse("space", "preview", "Planned", "", map[string]string{
		"page": "https://preview.invalid/page",
	})
	response.Configurations = append(response.Configurations, cm.PreviewEnvironmentConfiguration{
		URL:        "https://preview.invalid/unsupported",
		EntityType: cm.NewOptString("Entry"),
		EntityId:   cm.NewOptString("entry"),
		Enabled:    true,
	})

	state, responseDiagnostics, consistencyDiagnostics := ReconcilePreviewEnvironmentMutationResponse(
		t.Context(),
		response,
		plan,
		plan.PreviewEnvironmentIdentityModel,
	)

	require.False(t, responseDiagnostics.HasError())
	require.Len(t, responseDiagnostics.Warnings(), 1)
	require.True(t, consistencyDiagnostics.HasError())
	assert.Equal(t, []string{"content_type_configurations"}, attributeDiagnosticPaths(t, consistencyDiagnostics))
	assert.True(t, plan.ContentTypeConfigurations.Equal(state.ContentTypeConfigurations))
}

func previewEnvironmentConfigurationData(contentTypeID, url string, enabled bool) cm.PreviewEnvironmentConfigurationData {
	return cm.PreviewEnvironmentConfigurationData{
		URL:        url,
		EntityType: "ContentType",
		EntityId:   contentTypeID,
		Enabled:    enabled,
	}
}

func previewEnvironmentMutationModel(
	previewEnvironmentID string,
	description string,
	configurations map[string]string,
) PreviewEnvironmentModel {
	model := previewEnvironmentModel(configurations)
	model.ID = types.StringValue("space/" + previewEnvironmentID)
	model.SpaceID = types.StringValue("space")
	model.PreviewEnvironmentID = types.StringValue(previewEnvironmentID)
	model.Name = types.StringValue("Planned")
	model.Description = types.StringValue(description)

	return model
}

func previewEnvironmentResponse(
	spaceID string,
	previewEnvironmentID string,
	name string,
	description string,
	configurations map[string]string,
) cm.PreviewEnvironment {
	responseConfigurations := make([]cm.PreviewEnvironmentConfiguration, 0, len(configurations))
	for contentTypeID, url := range configurations {
		responseConfigurations = append(responseConfigurations, previewEnvironmentResponseConfiguration(contentTypeID, url, true))
	}

	return cm.PreviewEnvironment{
		Sys:            cm.NewPreviewEnvironmentSys(spaceID, previewEnvironmentID),
		Name:           name,
		Description:    description,
		Configurations: responseConfigurations,
	}
}

func previewEnvironmentResponseConfiguration(contentTypeID, url string, enabled bool) cm.PreviewEnvironmentConfiguration {
	return cm.PreviewEnvironmentConfiguration{
		URL:         url,
		EntityType:  cm.NewOptString("ContentType"),
		EntityId:    cm.NewOptString(contentTypeID),
		ContentType: cm.NewOptString(contentTypeID),
		Enabled:     enabled,
	}
}
