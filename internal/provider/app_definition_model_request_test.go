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

func validAppDefinitionRequestModel() AppDefinitionBaseModel {
	return AppDefinitionBaseModel{
		Name:     types.StringValue("App"),
		Src:      types.StringUnknown(),
		BundleID: types.StringUnknown(),
		Locations: []AppDefinitionLocationsItem{
			{
				Location: types.StringValue("entry-field"),
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
				NavigationItem: &AppDefinitionLocationNavigationItem{
					Name: types.StringValue("Entry"),
					Path: types.StringValue("/entry"),
				},
			},
		},
	}
}

func TestAppDefinitionRequestRejectsUnresolvedScalars(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		mutate       func(*AppDefinitionBaseModel)
		expectedPath string
	}{
		"unknown name": {
			mutate:       func(model *AppDefinitionBaseModel) { model.Name = types.StringUnknown() },
			expectedPath: "name",
		},
		"null name": {
			mutate:       func(model *AppDefinitionBaseModel) { model.Name = types.StringNull() },
			expectedPath: "name",
		},
		"unknown location": {
			mutate:       func(model *AppDefinitionBaseModel) { model.Locations[0].Location = types.StringUnknown() },
			expectedPath: "locations[0].location",
		},
		"null location": {
			mutate:       func(model *AppDefinitionBaseModel) { model.Locations[0].Location = types.StringNull() },
			expectedPath: "locations[0].location",
		},
		"unknown field type": {
			mutate:       func(model *AppDefinitionBaseModel) { model.Locations[0].FieldTypes[0].Type = types.StringUnknown() },
			expectedPath: "locations[0].field_types[0].type",
		},
		"null field type": {
			mutate:       func(model *AppDefinitionBaseModel) { model.Locations[0].FieldTypes[0].Type = types.StringNull() },
			expectedPath: "locations[0].field_types[0].type",
		},
		"unknown field link type": {
			mutate:       func(model *AppDefinitionBaseModel) { model.Locations[0].FieldTypes[0].LinkType = types.StringUnknown() },
			expectedPath: "locations[0].field_types[0].link_type",
		},
		"unknown item type": {
			mutate: func(model *AppDefinitionBaseModel) {
				model.Locations[0].FieldTypes[0].Items.Type = types.StringUnknown()
			},
			expectedPath: "locations[0].field_types[0].items.type",
		},
		"null item type": {
			mutate:       func(model *AppDefinitionBaseModel) { model.Locations[0].FieldTypes[0].Items.Type = types.StringNull() },
			expectedPath: "locations[0].field_types[0].items.type",
		},
		"unknown item link type": {
			mutate: func(model *AppDefinitionBaseModel) {
				model.Locations[0].FieldTypes[0].Items.LinkType = types.StringUnknown()
			},
			expectedPath: "locations[0].field_types[0].items.link_type",
		},
		"unknown navigation name": {
			mutate:       func(model *AppDefinitionBaseModel) { model.Locations[0].NavigationItem.Name = types.StringUnknown() },
			expectedPath: "locations[0].navigation_item.name",
		},
		"null navigation name": {
			mutate:       func(model *AppDefinitionBaseModel) { model.Locations[0].NavigationItem.Name = types.StringNull() },
			expectedPath: "locations[0].navigation_item.name",
		},
		"unknown navigation path": {
			mutate:       func(model *AppDefinitionBaseModel) { model.Locations[0].NavigationItem.Path = types.StringUnknown() },
			expectedPath: "locations[0].navigation_item.path",
		},
		"null navigation path": {
			mutate:       func(model *AppDefinitionBaseModel) { model.Locations[0].NavigationItem.Path = types.StringNull() },
			expectedPath: "locations[0].navigation_item.path",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := validAppDefinitionRequestModel()
			test.mutate(&model)

			actual, diags := model.ToAppDefinitionData(path.Empty())

			assert.Equal(t, cm.AppDefinitionData{}, actual)
			require.True(t, diags.HasError())
			assert.Equal(t, []string{test.expectedPath}, attributeDiagnosticPaths(t, diags))
		})
	}
}

func TestAppDefinitionRequestFailsAtomically(t *testing.T) {
	t.Parallel()

	model := validAppDefinitionRequestModel()
	model.Name = types.StringUnknown()
	model.Locations[0].Location = types.StringNull()
	model.Locations[0].FieldTypes[0].Items.LinkType = types.StringUnknown()
	model.Locations[0].NavigationItem.Path = types.StringUnknown()

	actual, diags := model.ToAppDefinitionData(path.Empty())

	assert.Equal(t, cm.AppDefinitionData{}, actual)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{
		"name",
		"locations[0].location",
		"locations[0].field_types[0].items.link_type",
		"locations[0].navigation_item.path",
	}, attributeDiagnosticPaths(t, diags))
}

func TestAppDefinitionRequestPreservesOptionalScalarPresence(t *testing.T) {
	t.Parallel()

	model := validAppDefinitionRequestModel()
	model.Name = types.StringValue("")
	model.Src = types.StringValue("")
	model.BundleID = types.StringValue("")
	model.Locations[0].FieldTypes[0].LinkType = types.StringValue("")
	model.Locations[0].FieldTypes[0].Items.LinkType = types.StringValue("")

	actual, diags := model.ToAppDefinitionData(path.Empty())

	require.False(t, diags.HasError(), diags.Errors())
	assert.Empty(t, actual.Name)
	assert.Empty(t, requireOptString(t, actual.Src))
	bundle, ok := actual.Bundle.Get()
	require.True(t, ok)
	assert.Empty(t, bundle.Sys.ID)
	assert.Empty(t, requireOptString(t, actual.Locations[0].FieldTypes[0].LinkType))
	items, ok := actual.Locations[0].FieldTypes[0].Items.Get()
	require.True(t, ok)
	assert.Empty(t, requireOptString(t, items.LinkType))
}

func TestAppDefinitionRequestOmitsOptionalComputedUnknownsAndNullLinks(t *testing.T) {
	t.Parallel()

	model := validAppDefinitionRequestModel()

	actual, diags := model.ToAppDefinitionData(path.Empty())

	require.False(t, diags.HasError(), diags.Errors())
	assert.False(t, actual.Src.IsSet())
	assert.False(t, actual.Bundle.IsSet())
	assert.False(t, actual.Locations[0].FieldTypes[0].LinkType.IsSet())
	items, ok := actual.Locations[0].FieldTypes[0].Items.Get()
	require.True(t, ok)
	assert.False(t, items.LinkType.IsSet())
}

func TestAppDefinitionRequestOmitsOptionalComputedNulls(t *testing.T) {
	t.Parallel()

	model := validAppDefinitionRequestModel()
	model.Src = types.StringNull()
	model.BundleID = types.StringNull()

	actual, diags := model.ToAppDefinitionData(path.Empty())

	require.False(t, diags.HasError(), diags.Errors())
	assert.False(t, actual.Src.IsSet())
	assert.False(t, actual.Bundle.IsSet())
}

func requireOptString(t *testing.T, value cm.OptString) string {
	t.Helper()

	actual, ok := value.Get()
	require.True(t, ok)

	return actual
}

func TestAppDefinitionParameterOptionsFailClosedWithExactPath(t *testing.T) {
	t.Parallel()

	model := AppDefinitionParameter{
		ID:      "parameter",
		Type:    "Enum",
		Name:    "Parameter",
		Default: jsontypes.NewNormalizedNull(),
		Options: NewTypedList([]jsontypes.Normalized{
			jsontypes.NewNormalizedValue(`"known"`),
			jsontypes.NewNormalizedNull(),
		}),
	}

	actual, diags := model.ToAppDefinitionParameter(
		path.Root("parameters").AtName("installation").AtListIndex(0),
	)

	assert.Zero(t, actual)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{"parameters.installation[0].options[1]"}, diagnosticPaths(t, diags))
}

func TestAppDefinitionParameterListsPreserveNullAndEmpty(t *testing.T) {
	t.Parallel()

	model := AppDefinitionBaseModel{
		Name: types.StringValue("App"),
		Parameters: &AppDefinitionParameters{
			Installation: []AppDefinitionParameter{},
			Instance:     nil,
		},
	}

	actual, diags := model.ToAppDefinitionData(path.Empty())

	require.False(t, diags.HasError())

	parameters, ok := actual.Parameters.Get()
	require.True(t, ok)
	require.NotNil(t, parameters.Installation)
	assert.Empty(t, parameters.Installation)
	assert.Nil(t, parameters.Instance)
}
