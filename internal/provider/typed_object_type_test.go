package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

func TestTypedObjectTypeFrameworkMethods(t *testing.T) {
	t.Parallel()

	type typedObjectTypeTestModel struct {
		Name types.String `tfsdk:"name"`
	}

	ctx := t.Context()
	objectType := TypedObject[typedObjectTypeTestModel]{}.CustomType(ctx)
	attributeTypes := map[string]attr.Type{"name": types.StringType}

	assert.IsType(t, TypedObject[typedObjectTypeTestModel]{}, objectType.ValueType(ctx))
	assert.True(t, objectType.WithAttributeTypes(attributeTypes).AttributeTypes()["name"].Equal(types.StringType))
	assertAttributePathStepType(t, objectType, tftypes.AttributeName("name"), types.StringType)
}
