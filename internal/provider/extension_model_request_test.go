package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validExtensionRequestModel() ExtensionModel {
	return ExtensionModel{
		Extension: &ExtensionModelExtension{
			Name:   types.StringValue("Extension"),
			Src:    types.StringUnknown(),
			SrcDoc: types.StringUnknown(),
			FieldTypes: []AppDefinitionLocationFieldTypesItem{
				{
					Type:     types.StringValue("Array"),
					LinkType: types.StringNull(),
					Items: &AppDefinitionLocationFieldTypeItemsItem{
						Type:     types.StringValue("Link"),
						LinkType: types.StringNull(),
					},
				},
			},
			Sidebar: types.BoolValue(false),
		},
		Parameters: jsontypes.NewNormalizedUnknown(),
	}
}

func TestExtensionRequestRejectsUnresolvedScalars(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		mutate       func(*ExtensionModel)
		expectedPath string
	}{
		"unknown name": {
			mutate:       func(model *ExtensionModel) { model.Extension.Name = types.StringUnknown() },
			expectedPath: "extension.name",
		},
		"null name": {
			mutate:       func(model *ExtensionModel) { model.Extension.Name = types.StringNull() },
			expectedPath: "extension.name",
		},
		"unknown field type": {
			mutate:       func(model *ExtensionModel) { model.Extension.FieldTypes[0].Type = types.StringUnknown() },
			expectedPath: "extension.field_types[0].type",
		},
		"null field type": {
			mutate:       func(model *ExtensionModel) { model.Extension.FieldTypes[0].Type = types.StringNull() },
			expectedPath: "extension.field_types[0].type",
		},
		"unknown field link type": {
			mutate:       func(model *ExtensionModel) { model.Extension.FieldTypes[0].LinkType = types.StringUnknown() },
			expectedPath: "extension.field_types[0].link_type",
		},
		"unknown item type": {
			mutate:       func(model *ExtensionModel) { model.Extension.FieldTypes[0].Items.Type = types.StringUnknown() },
			expectedPath: "extension.field_types[0].items.type",
		},
		"null item type": {
			mutate:       func(model *ExtensionModel) { model.Extension.FieldTypes[0].Items.Type = types.StringNull() },
			expectedPath: "extension.field_types[0].items.type",
		},
		"unknown item link type": {
			mutate:       func(model *ExtensionModel) { model.Extension.FieldTypes[0].Items.LinkType = types.StringUnknown() },
			expectedPath: "extension.field_types[0].items.link_type",
		},
		"unknown sidebar": {
			mutate:       func(model *ExtensionModel) { model.Extension.Sidebar = types.BoolUnknown() },
			expectedPath: "extension.sidebar",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := validExtensionRequestModel()
			test.mutate(&model)

			actual, diags := model.ToExtensionData(t.Context(), path.Empty())

			assert.Equal(t, cm.ExtensionData{}, actual)
			require.True(t, diags.HasError())
			assert.Equal(t, []string{test.expectedPath}, attributeDiagnosticPaths(t, diags))
		})
	}
}

func TestExtensionRequestFailsAtomically(t *testing.T) {
	t.Parallel()

	model := validExtensionRequestModel()
	model.Extension.Name = types.StringUnknown()
	model.Extension.FieldTypes[0].LinkType = types.StringUnknown()
	model.Extension.FieldTypes[0].Items.Type = types.StringNull()
	model.Extension.Sidebar = types.BoolUnknown()

	actual, diags := model.ToExtensionData(t.Context(), path.Empty())

	assert.Equal(t, cm.ExtensionData{}, actual)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{
		"extension.name",
		"extension.sidebar",
		"extension.field_types[0].link_type",
		"extension.field_types[0].items.type",
	}, attributeDiagnosticPaths(t, diags))
}

func TestExtensionRequestPreservesOptionalScalarPresence(t *testing.T) {
	t.Parallel()

	model := validExtensionRequestModel()
	model.Extension.Name = types.StringValue("")
	model.Extension.Src = types.StringValue("")
	model.Extension.SrcDoc = types.StringValue("")
	model.Extension.FieldTypes[0].LinkType = types.StringValue("")
	model.Extension.FieldTypes[0].Items.LinkType = types.StringValue("")
	model.Extension.Sidebar = types.BoolValue(false)

	actual, diags := model.ToExtensionData(t.Context(), path.Empty())

	require.False(t, diags.HasError(), diags.Errors())
	assert.Empty(t, actual.Extension.Name)
	assert.False(t, actual.Extension.Src.IsSet())
	assert.False(t, actual.Extension.Srcdoc.IsSet())
	assert.Empty(t, requireOptString(t, actual.Extension.FieldTypes[0].LinkType))
	items, ok := actual.Extension.FieldTypes[0].Items.Get()
	require.True(t, ok)
	assert.Empty(t, requireOptString(t, items.LinkType))

	sidebar, ok := actual.Extension.Sidebar.Get()
	require.True(t, ok)
	assert.False(t, sidebar)
}

func TestExtensionRequestOmitsSchemaConsistentNulls(t *testing.T) {
	t.Parallel()

	model := validExtensionRequestModel()
	model.Extension.Src = types.StringNull()
	model.Extension.SrcDoc = types.StringNull()
	model.Extension.Sidebar = types.BoolNull()

	actual, diags := model.ToExtensionData(t.Context(), path.Empty())

	require.False(t, diags.HasError(), diags.Errors())
	assert.False(t, actual.Extension.Src.IsSet())
	assert.False(t, actual.Extension.Srcdoc.IsSet())
	assert.False(t, actual.Extension.Sidebar.IsSet())
	assert.False(t, actual.Extension.FieldTypes[0].LinkType.IsSet())
	items, ok := actual.Extension.FieldTypes[0].Items.Get()
	require.True(t, ok)
	assert.False(t, items.LinkType.IsSet())
}

func TestExtensionRequestPreservesKnownSources(t *testing.T) {
	t.Parallel()

	model := validExtensionRequestModel()
	model.Extension.Src = types.StringValue("https://example.com")
	model.Extension.SrcDoc = types.StringValue("<html></html>")
	model.Parameters = jsontypes.NewNormalizedValue(`{"known":true}`)

	actual, diags := model.ToExtensionData(t.Context(), path.Empty())

	require.False(t, diags.HasError(), diags.Errors())
	assert.Equal(t, "https://example.com", requireOptString(t, actual.Extension.Src))
	assert.Equal(t, "<html></html>", requireOptString(t, actual.Extension.Srcdoc))
	assert.JSONEq(t, `{"known":true}`, string(actual.Parameters))
}

func TestExtensionParameterListsFailClosed(t *testing.T) {
	t.Parallel()

	model := ExtensionModel{
		Extension: &ExtensionModelExtension{
			Name: types.StringValue("Extension"),
			Parameters: &AppDefinitionParameters{
				Installation: []AppDefinitionParameter{
					{
						ID:      "known",
						Default: jsontypes.NewNormalizedNull(),
						Options: NewTypedListNull[jsontypes.Normalized](),
					},
				},
				Instance: []AppDefinitionParameter{
					{
						ID:      "unresolved",
						Default: jsontypes.NewNormalizedNull(),
						Options: NewTypedList([]jsontypes.Normalized{jsontypes.NewNormalizedUnknown()}),
					},
				},
			},
		},
		Parameters: jsontypes.NewNormalizedNull(),
	}

	actual, diags := model.ToExtensionData(t.Context(), path.Empty())

	assert.Equal(t, cm.ExtensionData{}, actual)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{"extension.parameters.instance[0].options[0]"}, attributeDiagnosticPaths(t, diags))
}

func TestExtensionParameterListsPreserveNilAndEmpty(t *testing.T) {
	t.Parallel()

	model := ExtensionModel{
		Extension: &ExtensionModelExtension{
			Name: types.StringValue("Extension"),
			Parameters: &AppDefinitionParameters{
				Installation: []AppDefinitionParameter{},
				Instance:     nil,
			},
		},
		Parameters: jsontypes.NewNormalizedNull(),
	}

	actual, diags := model.ToExtensionData(t.Context(), path.Empty())

	require.False(t, diags.HasError(), diags.Errors())

	parameters, ok := actual.Extension.Parameters.Get()
	require.True(t, ok)
	require.NotNil(t, parameters.Installation)
	assert.Empty(t, parameters.Installation)
	assert.Nil(t, parameters.Instance)
}

func TestExtensionOptionalComputedValuesCanRemainUnknown(t *testing.T) {
	t.Parallel()

	model := ExtensionModel{
		Extension: &ExtensionModelExtension{
			Name:   types.StringValue("Extension"),
			Src:    types.StringUnknown(),
			SrcDoc: types.StringUnknown(),
		},
		Parameters: jsontypes.NewNormalizedUnknown(),
	}

	actual, diags := model.ToExtensionData(t.Context(), path.Empty())

	require.False(t, diags.HasError(), diags.Errors())
	assert.False(t, actual.Extension.Src.IsSet())
	assert.False(t, actual.Extension.Srcdoc.IsSet())
	assert.Nil(t, actual.Parameters)
}
