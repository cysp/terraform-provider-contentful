package terraformpluginframeworkreflection_test

import (
	"testing"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValueToTerraformValueNull(t *testing.T) {
	t.Parallel()

	actual, err := tpfr.ValueToTerraformValue(t.Context(), testValue{}, attr.ValueStateNull)
	require.NoError(t, err)

	assert.True(t, actual.IsKnown())
	assert.True(t, actual.IsNull())
}

func TestValueToTerraformValueUnknown(t *testing.T) {
	t.Parallel()

	actual, err := tpfr.ValueToTerraformValue(t.Context(), testValue{}, attr.ValueStateUnknown)
	require.NoError(t, err)

	assert.False(t, actual.IsKnown())
	assert.False(t, actual.IsNull())
}

func TestValueToTerraformValueUnexpectedState(t *testing.T) {
	t.Parallel()

	actual, err := tpfr.ValueToTerraformValue(t.Context(), testValue{}, 0xff)
	require.Error(t, err)

	assert.True(t, actual.IsKnown())
	assert.True(t, actual.IsNull())

	assert.Equal(t, "unexpected value state: 0xff", err.Error())

	var uvserr tpfr.UnexpectedValueStateError

	require.ErrorAs(t, err, &uvserr)
	assert.Equal(t, attr.ValueState(0xff), uvserr.ValueState)
}
