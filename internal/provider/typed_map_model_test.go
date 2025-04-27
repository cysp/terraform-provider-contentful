package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypedMapInModels(t *testing.T) {
	ctx := context.Background()

	// Test model with TypedMap field
	type TestModel struct {
		ID           types.String                      `tfsdk:"id"`
		Name         types.String                      `tfsdk:"name"`
		Permissions  TypedMap[TypedList[types.String]] `tfsdk:"permissions"`
	}

	// Create permissions data
	actionsList1, diags := NewTypedList(ctx, []types.String{
		types.StringValue("create"),
		types.StringValue("read"),
	})
	require.False(t, diags.HasError())

	actionsList2, diags := NewTypedList(ctx, []types.String{
		types.StringValue("all"),
	})
	require.False(t, diags.HasError())

	elements := map[string]TypedList[types.String]{
		"entry":   actionsList1,
		"content": actionsList2,
	}

	permissionsMap, diags := NewTypedMap(ctx, elements)
	require.False(t, diags.HasError())

	// Create model instance
	model := TestModel{
		ID:          types.StringValue("test-id"),
		Name:        types.StringValue("Test Model"),
		Permissions: permissionsMap,
	}

	// Test accessing TypedMap data from model
	assert.Equal(t, attr.ValueStateKnown, model.Permissions.state)
	assert.Equal(t, 2, len(model.Permissions.Elements()))

	// Test accessing list elements within the map
	entryPermissions := model.Permissions.Elements()["entry"]
	assert.Equal(t, 2, len(entryPermissions.Elements()))

	contentPermissions := model.Permissions.Elements()["content"]
	assert.Equal(t, 1, len(contentPermissions.Elements()))
	assert.Equal(t, "all", contentPermissions.Elements()[0].ValueString())
}

func TestCustomTypeIntegration(t *testing.T) {
	ctx := context.Background()

	// Test that TypedMap can be used as a CustomType
	mapType := NewTypedMapNull[TypedList[types.String]](ctx).CustomType(ctx)
	assert.NotNil(t, mapType)

	// Verify Type and ValueType methods
	valueTypeObj := mapType.ValueType(ctx)
	assert.IsType(t, TypedMap[TypedList[types.String]]{}, valueTypeObj)

	// Test element type functionality
	elementType := mapType.ElementType()
	assert.NotNil(t, elementType)

	// Test WithElementType
	newMapType := mapType.WithElementType(types.StringType)
	assert.NotNil(t, newMapType)

	// Verify Equal implementation
	sameMapType := NewTypedMapNull[TypedList[types.String]](ctx).CustomType(ctx)
	assert.True(t, mapType.Equal(sameMapType))

	differentMapType := NewTypedMapNull[types.String](ctx).CustomType(ctx)
	assert.False(t, mapType.Equal(differentMapType))
}