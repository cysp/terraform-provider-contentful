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

func TestTypedObjectTypeEqual(t *testing.T) {
	t.Parallel()

	type typedObjectTypeTestModel struct {
		Name types.String `tfsdk:"name"`
	}

	type otherTypedObjectTypeTestModel struct {
		Name types.String `tfsdk:"name"`
	}

	objectType := TypedObject[typedObjectTypeTestModel]{}.CustomType(t.Context())
	nameType := objectType.WithAttributeTypes(map[string]attr.Type{
		"name": types.StringType,
	})
	equalNameType := objectType.WithAttributeTypes(map[string]attr.Type{
		"name": types.StringType,
	})
	missingNameType := objectType.WithAttributeTypes(map[string]attr.Type{})
	extraAttributeType := objectType.WithAttributeTypes(map[string]attr.Type{
		"name":    types.StringType,
		"enabled": types.BoolType,
	})
	differentNameType := objectType.WithAttributeTypes(map[string]attr.Type{
		"name": types.BoolType,
	})

	assert.True(t, nameType.Equal(equalNameType))
	assert.True(t, equalNameType.Equal(nameType))

	testcases := map[string]struct {
		left  attr.Type
		right attr.Type
	}{
		"missing attribute": {
			left:  nameType,
			right: missingNameType,
		},
		"extra attribute": {
			left:  nameType,
			right: extraAttributeType,
		},
		"different attribute type": {
			left:  nameType,
			right: differentNameType,
		},
		"different model with identical attributes": {
			left:  objectType,
			right: TypedObject[otherTypedObjectTypeTestModel]{}.CustomType(t.Context()),
		},
		"different Go type with identical attributes": {
			left:  objectType,
			right: types.ObjectType{AttrTypes: map[string]attr.Type{"name": types.StringType}},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.False(t, testcase.left.Equal(testcase.right))
			assert.False(t, testcase.right.Equal(testcase.left))
		})
	}
}
