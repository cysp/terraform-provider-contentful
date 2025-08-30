package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestTypedMapEqual(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	map1 := DiagsNoErrorsMust(NewTypedMap(ctx, map[string]types.String{
		"key1": types.StringValue("value1"),
		"key2": types.StringValue("value2"),
	}))

	map2 := DiagsNoErrorsMust(NewTypedMap(ctx, map[string]types.String{
		"key1": types.StringValue("value1"),
		"key2": types.StringValue("value2"),
	}))

	mapDifferentValues := DiagsNoErrorsMust(NewTypedMap(ctx, map[string]types.String{
		"key1": types.StringValue("different"),
		"key2": types.StringValue("value2"),
	}))

	mapDifferentKeys := DiagsNoErrorsMust(NewTypedMap(ctx, map[string]types.String{
		"key1": types.StringValue("value1"),
		"key3": types.StringValue("value3"),
	}))

	mapDifferentLength := DiagsNoErrorsMust(NewTypedMap(ctx, map[string]types.String{
		"key1": types.StringValue("value1"),
		"key2": types.StringValue("value2"),
		"key3": types.StringValue("value3"),
	}))

	testcases := map[string]struct {
		map1     TypedMap[types.String]
		map2     attr.Value
		expected bool
	}{
		"equal maps": {
			map1:     map1,
			map2:     map2,
			expected: true,
		},
		"different values": {
			map1:     map1,
			map2:     mapDifferentValues,
			expected: false,
		},
		"different keys": {
			map1:     map1,
			map2:     mapDifferentKeys,
			expected: false,
		},
		"different length": {
			map1:     map1,
			map2:     mapDifferentLength,
			expected: false,
		},
		"null != known": {
			map1:     NewTypedMapNull[types.String](ctx),
			map2:     map1,
			expected: false,
		},
		"unknown != known": {
			map1:     NewTypedMapUnknown[types.String](ctx),
			map2:     map1,
			expected: false,
		},
		"null == null": {
			map1:     NewTypedMapNull[types.String](ctx),
			map2:     NewTypedMapNull[types.String](ctx),
			expected: true,
		},
		"unknown == unknown": {
			map1:     NewTypedMapUnknown[types.String](ctx),
			map2:     NewTypedMapUnknown[types.String](ctx),
			expected: true,
		},
		"different type": {
			map1:     map1,
			map2:     types.StringValue("string"),
			expected: false,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testcase.expected, testcase.map1.Equal(testcase.map2))
		})
	}
}
