package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		t.Context(),
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

	actual, diags := model.ToAppDefinitionData(t.Context(), path.Empty())

	require.False(t, diags.HasError())

	parameters, ok := actual.Parameters.Get()
	require.True(t, ok)
	require.NotNil(t, parameters.Installation)
	assert.Empty(t, parameters.Installation)
	assert.Nil(t, parameters.Instance)
}
