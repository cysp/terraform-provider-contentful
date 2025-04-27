package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRolePermissionsConversion(t *testing.T) {
	ctx := context.Background()
	testPath := path.Root("permissions")

	// Test cases
	tests := []struct {
		name           string
		setupMap       func(context.Context) TypedMap[TypedList[types.String]]
		expectedLength int
		expectError    bool
	}{
		{
			name: "valid permissions map",
			setupMap: func(ctx context.Context) TypedMap[TypedList[types.String]] {
				actions1, diags := NewTypedList(ctx, []types.String{
					types.StringValue("create"),
					types.StringValue("read"),
				})
				require.False(t, diags.HasError())

				actions2, diags := NewTypedList(ctx, []types.String{
					types.StringValue("all"),
				})
				require.False(t, diags.HasError())

				elements := map[string]TypedList[types.String]{
					"entry":   actions1,
					"content": actions2,
				}

				result, diags := NewTypedMap(ctx, elements)
				require.False(t, diags.HasError())
				return result
			},
			expectedLength: 2,
			expectError:    false,
		},
		{
			name: "unknown permissions map",
			setupMap: func(ctx context.Context) TypedMap[TypedList[types.String]] {
				return NewTypedMapUnknown[TypedList[types.String]](ctx)
			},
			expectedLength: 0,
			expectError:    false,
		},
		{
			name: "null permissions map",
			setupMap: func(ctx context.Context) TypedMap[TypedList[types.String]] {
				return NewTypedMapNull[TypedList[types.String]](ctx)
			},
			expectedLength: 0,
			expectError:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			permissionsMap := test.setupMap(ctx)

			// Test permission map access through Elements()
			if permissionsMap.state == attr.ValueStateKnown {
				elements := permissionsMap.Elements()
				assert.Equal(t, test.expectedLength, len(elements))
			}

			// Test TypedMap to/from MapValue conversion
			if permissionsMap.state == attr.ValueStateKnown {
				mapValue, diags := permissionsMap.ToMapValue(ctx)
				require.False(t, diags.HasError())

				mapType := permissionsMap.Type(ctx)
				mapTyped, ok := mapType.(TypedMapType[TypedList[types.String]])
				require.True(t, ok)

				converted, diags := mapTyped.ValueFromMap(ctx, mapValue)
				require.False(t, diags.HasError())

				reconverted, ok := converted.(TypedMap[TypedList[types.String]])
				require.True(t, ok)

				assert.True(t, permissionsMap.Equal(reconverted))
			}
		})
	}
}

func TestRoleFieldsPermissionsConversion(t *testing.T) {
	ctx := context.Background()
	testPath := path.Root("permissions")

	// Setup a test permissions map
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
	
	// Test conversion to RoleFieldsPermissions
	roleFieldsPermissions, diags := ToRoleFieldsPermissions(ctx, testPath, permissionsMap)
	require.False(t, diags.HasError())
	
	// Verify conversion results
	assert.Equal(t, 2, len(roleFieldsPermissions))
	
	// Verify conversion of "all" permission
	contentPermission, ok := roleFieldsPermissions["content"]
	require.True(t, ok)
	assert.Equal(t, "all", contentPermission.Type)
	
	// Verify conversion of regular permissions list
	entryPermission, ok := roleFieldsPermissions["entry"]
	require.True(t, ok)
	assert.Equal(t, "array", entryPermission.Type)
	assert.Contains(t, entryPermission.Actions, "create")
	assert.Contains(t, entryPermission.Actions, "read")
}