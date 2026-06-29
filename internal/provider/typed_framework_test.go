package provider_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type typedElementType interface {
	WithElementType(typ attr.Type) attr.TypeWithElementType
	ElementTypeWithContext(ctx context.Context) attr.Type
}

type terraform5PathStepper interface {
	ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error)
}

func assertStringElementType(t *testing.T, typ typedElementType) {
	t.Helper()

	assert.True(t, typ.WithElementType(types.StringType).ElementType().Equal(types.StringType))
	assert.True(t, typ.ElementTypeWithContext(t.Context()).Equal(types.StringType))
}

func assertAttributePathStepType(t *testing.T, typ terraform5PathStepper, step tftypes.AttributePathStep, expected attr.Type) {
	t.Helper()

	stepType, err := typ.ApplyTerraform5AttributePathStep(step)
	require.NoError(t, err)

	stepAttrType, ok := stepType.(attr.Type)
	require.True(t, ok)

	assert.True(t, stepAttrType.Equal(expected))
}
