package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppDefinitionParameterToRequestRejectsUnresolvedJSON(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		model               AppDefinitionParameter
		expectedDiagnostics []string
	}{
		"unknown default": {
			model: AppDefinitionParameter{
				Default: jsontypes.NewNormalizedUnknown(),
				Options: NewTypedListNull[jsontypes.Normalized](),
			},
			expectedDiagnostics: []string{"parameters.installation[0].default"},
		},
		"unknown options": {
			model: AppDefinitionParameter{
				Default: jsontypes.NewNormalizedNull(),
				Options: NewTypedListUnknown[jsontypes.Normalized](),
			},
			expectedDiagnostics: []string{"parameters.installation[0].options"},
		},
		"unknown option child": {
			model: AppDefinitionParameter{
				Default: jsontypes.NewNormalizedNull(),
				Options: NewTypedList([]jsontypes.Normalized{
					jsontypes.NewNormalizedValue(`"known"`),
					jsontypes.NewNormalizedUnknown(),
				}),
			},
			expectedDiagnostics: []string{"parameters.installation[0].options[1]"},
		},
		"null option child": {
			model: AppDefinitionParameter{
				Default: jsontypes.NewNormalizedNull(),
				Options: NewTypedList([]jsontypes.Normalized{
					jsontypes.NewNormalizedValue(`"known"`),
					jsontypes.NewNormalizedNull(),
				}),
			},
			expectedDiagnostics: []string{"parameters.installation[0].options[1]"},
		},
		"multiple unresolved values": {
			model: AppDefinitionParameter{
				Default: jsontypes.NewNormalizedUnknown(),
				Options: NewTypedList([]jsontypes.Normalized{
					jsontypes.NewNormalizedUnknown(),
					jsontypes.NewNormalizedNull(),
				}),
			},
			expectedDiagnostics: []string{
				"parameters.installation[0].default",
				"parameters.installation[0].options[0]",
				"parameters.installation[0].options[1]",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := test.model.ToAppDefinitionParameter(
				path.Root("parameters").AtName("installation").AtListIndex(0),
			)

			assert.Equal(t, cm.AppDefinitionParameter{}, actual)
			require.True(t, diags.HasError())
			assert.Equal(t, test.expectedDiagnostics, attributeDiagnosticPaths(t, diags))
		})
	}
}

func TestAppDefinitionParameterToRequestPreservesJSONPresence(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		model                  AppDefinitionParameter
		expectedDefault        jx.Raw
		expectedOptions        []jx.Raw
		expectedOptionsPresent bool
	}{
		"absent": {
			model: AppDefinitionParameter{
				Default: jsontypes.NewNormalizedNull(),
				Options: NewTypedListNull[jsontypes.Normalized](),
			},
		},
		"known JSON null": {
			model: AppDefinitionParameter{
				Default: jsontypes.NewNormalizedValue("null"),
				Options: NewTypedList([]jsontypes.Normalized{jsontypes.NewNormalizedValue("null")}),
			},
			expectedDefault:        jx.Raw("null"),
			expectedOptions:        []jx.Raw{jx.Raw("null")},
			expectedOptionsPresent: true,
		},
		"known empty options": {
			model: AppDefinitionParameter{
				Default: jsontypes.NewNormalizedNull(),
				Options: NewTypedList([]jsontypes.Normalized{}),
			},
			expectedOptions:        []jx.Raw{},
			expectedOptionsPresent: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := test.model.ToAppDefinitionParameter(path.Root("parameter"))

			require.False(t, diags.HasError(), diags.Errors())
			assert.Equal(t, test.expectedDefault, actual.Default)
			assert.Equal(t, test.expectedOptions, actual.Options)
			assert.Equal(t, test.expectedOptionsPresent, actual.Options != nil)
		})
	}
}

func TestAppDefinitionParameterListsFailClosed(t *testing.T) {
	t.Parallel()

	model := AppDefinitionBaseModel{
		Name: types.StringValue("App"),
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
					Default: jsontypes.NewNormalizedUnknown(),
					Options: NewTypedListNull[jsontypes.Normalized](),
				},
			},
		},
	}

	actual, diags := model.ToAppDefinitionData(AppDefinitionBaseModel{}, path.Empty())

	assert.Equal(t, cm.AppDefinitionData{}, actual)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{"parameters.instance[0].default"}, attributeDiagnosticPaths(t, diags))
}

func TestAppDefinitionParameterListsPreserveNilAndEmpty(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		parameters         *AppDefinitionParameters
		expectParameters   bool
		expectInstallation bool
		expectInstance     bool
	}{
		"absent parameters": {},
		"empty installation and nil instance": {
			parameters: &AppDefinitionParameters{
				Installation: []AppDefinitionParameter{},
				Instance:     nil,
			},
			expectParameters:   true,
			expectInstallation: true,
		},
		"nil installation and empty instance": {
			parameters: &AppDefinitionParameters{
				Installation: nil,
				Instance:     []AppDefinitionParameter{},
			},
			expectParameters: true,
			expectInstance:   true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := AppDefinitionBaseModel{
				Name:       types.StringValue("App"),
				Parameters: test.parameters,
			}

			actual, diags := model.ToAppDefinitionData(AppDefinitionBaseModel{}, path.Empty())

			require.False(t, diags.HasError(), diags.Errors())

			parameters, ok := actual.Parameters.Get()
			assert.Equal(t, test.expectParameters, ok)

			if !ok {
				return
			}

			assert.Equal(t, test.expectInstallation, parameters.Installation != nil)
			assert.Equal(t, test.expectInstance, parameters.Instance != nil)
		})
	}
}
