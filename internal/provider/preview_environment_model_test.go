package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
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

func TestPreviewEnvironmentModelResponseRejectsInvalidConfigurations(t *testing.T) {
	t.Parallel()

	tests := map[string][]cm.PreviewEnvironmentConfiguration{
		"missing identity": {
			{URL: "https://preview.invalid/page", Enabled: true},
		},
		"conflicting aliases": {
			{
				URL:         "https://preview.invalid/page",
				EntityType:  cm.NewOptString("ContentType"),
				EntityId:    cm.NewOptString("page"),
				ContentType: cm.NewOptString("article"),
				Enabled:     true,
			},
		},
		"unsupported enabled entity": {
			{
				URL:        "https://preview.invalid/page",
				EntityType: cm.NewOptString("Entry"),
				EntityId:   cm.NewOptString("page"),
				Enabled:    true,
			},
		},
		"unsupported disabled entity": {
			{
				URL:        "https://preview.invalid/page",
				EntityType: cm.NewOptString("Entry"),
				EntityId:   cm.NewOptString("page"),
				Enabled:    false,
			},
		},
		"duplicate identity": {
			{
				URL:        "https://preview.invalid/page",
				EntityType: cm.NewOptString("ContentType"),
				EntityId:   cm.NewOptString("page"),
				Enabled:    true,
			},
			{
				URL:        "https://preview.invalid/page-disabled",
				EntityType: cm.NewOptString("ContentType"),
				EntityId:   cm.NewOptString("page"),
				Enabled:    false,
			},
		},
	}

	for name, configurations := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, diagnostics := NewPreviewEnvironmentModelFromResponse(t.Context(), cm.PreviewEnvironment{
				Sys:            cm.NewPreviewEnvironmentSys("space", "preview"),
				Name:           "Preview",
				Description:    "",
				Configurations: configurations,
			})
			require.True(t, diagnostics.HasError())
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

func TestValidatePreviewEnvironmentUpdateResponse(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		state       map[string]string
		plan        map[string]string
		response    map[string]string
		expectError bool
	}{
		"changed value applied with unrelated remote value": {
			state: map[string]string{"page": "https://preview.invalid/page"},
			plan:  map[string]string{"page": "https://preview.invalid/page-new"},
			response: map[string]string{
				"page":   "https://preview.invalid/page-new",
				"author": "https://preview.invalid/author",
			},
		},
		"changed value not applied": {
			state:       map[string]string{"page": "https://preview.invalid/page"},
			plan:        map[string]string{"page": "https://preview.invalid/page-new"},
			response:    map[string]string{"page": "https://preview.invalid/page"},
			expectError: true,
		},
		"removal applied": {
			state:    map[string]string{"page": "https://preview.invalid/page"},
			plan:     map[string]string{},
			response: map[string]string{},
		},
		"removal not applied": {
			state:       map[string]string{"page": "https://preview.invalid/page"},
			plan:        map[string]string{},
			response:    map[string]string{"page": "https://preview.invalid/page"},
			expectError: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			state := previewEnvironmentModel(test.state)
			plan := previewEnvironmentModel(test.plan)
			response := previewEnvironmentModel(test.response)
			diagnostics := ValidatePreviewEnvironmentUpdateResponse(
				t.Context(),
				path.Empty(),
				&state,
				&plan,
				&response,
			)
			assert.Equal(t, test.expectError, diagnostics.HasError())
		})
	}
}

func previewEnvironmentConfigurationData(contentTypeID, url string, enabled bool) cm.PreviewEnvironmentConfigurationData {
	return cm.PreviewEnvironmentConfigurationData{
		URL:        url,
		EntityType: "ContentType",
		EntityId:   contentTypeID,
		Enabled:    enabled,
	}
}
