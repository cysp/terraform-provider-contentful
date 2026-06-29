package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

func TestTypedMapTypeFrameworkMethods(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	mapType := TypedMap[types.String]{}.CustomType(ctx)

	assert.IsType(t, TypedMap[types.String]{}, mapType.ValueType(ctx))
	assert.Equal(t, "TypedMap[basetypes.StringValue]", mapType.String())
	assertStringElementType(t, mapType)
	assertAttributePathStepType(t, mapType, tftypes.ElementKeyString("key"), types.StringType)
}
