package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypedMapEqual(t *testing.T) {
	ctx := context.Background()

	map1, diags := NewTypedMap(ctx, map[string]types.String{
		"key1": types.StringValue("value1"),
		"key2": types.StringValue("value2"),
	})
	require.False(t, diags.HasError())

	map2, diags := NewTypedMap(ctx, map[string]types.String{
		"key1": types.StringValue("value1"),
		"key2": types.StringValue("value2"),
	})
	require.False(t, diags.HasError())

	mapDifferentValues, diags := NewTypedMap(ctx, map[string]types.String{
		"key1": types.StringValue("different"),
		"key2": types.StringValue("value2"),
	})
	require.False(t, diags.HasError())

	mapDifferentKeys, diags := NewTypedMap(ctx, map[string]types.String{
		"key1": types.StringValue("value1"),
		"key3": types.StringValue("value3"),
	})
	require.False(t, diags.HasError())

	mapDifferentLength, diags := NewTypedMap(ctx, map[string]types.String{
		"key1": types.StringValue("value1"),
		"key2": types.StringValue("value2"),
		"key3": types.StringValue("value3"),
	})
	require.False(t, diags.HasError())

	nullMap := NewTypedMapNull[types.String](ctx)
	unknownMap := NewTypedMapUnknown[types.String](ctx)

	testCases := []struct {
		name     string
		map1     TypedMap[types.String]
		map2     attr.Value
		expected bool
	}{
		{
			name:     "equal maps",
			map1:     map1,
			map2:     map2,
			expected: true,
		},
		{
			name:     "different values",
			map1:     map1,
			map2:     mapDifferentValues,
			expected: false,
		},
		{
			name:     "different keys",
			map1:     map1,
			map2:     mapDifferentKeys,
			expected: false,
		},
		{
			name:     "different length",
			map1:     map1,
			map2:     mapDifferentLength,
			expected: false,
		},
		{
			name:     "null vs known",
			map1:     nullMap,
			map2:     map1,
			expected: false,
		},
		{
			name:     "unknown vs known",
			map1:     unknownMap,
			map2:     map1,
			expected: false,
		},
		{
			name:     "null and null",
			map1:     nullMap,
			map2:     NewTypedMapNull[types.String](ctx),
			expected: true,
		},
		{
			name:     "unknown and unknown",
			map1:     unknownMap,
			map2:     NewTypedMapUnknown[types.String](ctx),
			expected: true,
		},
		{
			name:     "different type",
			map1:     map1,
			map2:     types.StringValue("not a map"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.map1.Equal(tc.map2))
		})
	}
}

func TestTypedMapValueFromTerraform(t *testing.T) {
	ctx := context.Background()
	
	// Define test cases
	tests := []struct {
		name          string
		value         tftypes.Value
		expected      map[string]types.String
		expectedState attr.ValueState
		expectError   bool
	}{
		{
			name: "string map",
			value: tftypes.NewValue(
				tftypes.Map{ElementType: tftypes.String},
				map[string]tftypes.Value{
					"key1": tftypes.NewValue(tftypes.String, "value1"),
					"key2": tftypes.NewValue(tftypes.String, "value2"),
				},
			),
			expected: map[string]types.String{
				"key1": types.StringValue("value1"),
				"key2": types.StringValue("value2"),
			},
			expectedState: attr.ValueStateKnown,
			expectError:   false,
		},
		{
			name:          "null map",
			value:         tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
			expected:      map[string]types.String{},
			expectedState: attr.ValueStateNull,
			expectError:   false,
		},
		{
			name:          "unknown map",
			value:         tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, tftypes.UnknownValue),
			expected:      map[string]types.String{},
			expectedState: attr.ValueStateUnknown,
			expectError:   false,
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create the map type
			mapType := TypedMapType[types.String]{elementType: types.StringType}
			
			// Convert from Terraform value
			got, err := mapType.ValueFromTerraform(ctx, test.value)
			if test.expectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			
			// Check the result
			typedMap, ok := got.(TypedMap[types.String])
			require.True(t, ok, "Expected TypedMap[types.String], got %T", got)
			
			assert.Equal(t, test.expectedState, typedMap.state)
			if test.expectedState == attr.ValueStateKnown {
				assert.Equal(t, test.expected, typedMap.elements)
			}
		})
	}
}

func TestPermissionsMapFunctions(t *testing.T) {
	ctx := context.Background()
	testPath := path.Root("test")
	
	// Test creating a map from elements
	elements := map[string]TypedList[types.String]{}
	
	// Create a typed list for a test entry
	stringList1, diags := NewTypedList(ctx, []types.String{
		types.StringValue("create"),
		types.StringValue("read"),
	})
	require.False(t, diags.HasError())
	
	stringList2, diags := NewTypedList(ctx, []types.String{
		types.StringValue("all"),
	})
	require.False(t, diags.HasError())
	
	elements["entry"] = stringList1
	elements["content"] = stringList2
	
	// Create a typed map
	typedMap, diags := NewTypedMap(ctx, elements)
	require.False(t, diags.HasError())
	
	// Test ToMapValue
	mapValue, diags := typedMap.ToMapValue(ctx)
	require.False(t, diags.HasError())
	require.NotNil(t, mapValue)
	
	// Test Elements() access
	retrievedElements := typedMap.Elements()
	assert.Equal(t, elements, retrievedElements)
}