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

	map1 := NewTypedMap(map[string]types.String{
		"key1": types.StringValue("value1"),
		"key2": types.StringValue("value2"),
	})

	map2 := NewTypedMap(map[string]types.String{
		"key1": types.StringValue("value1"),
		"key2": types.StringValue("value2"),
	})

	mapDifferentValues := NewTypedMap(map[string]types.String{
		"key1": types.StringValue("different"),
		"key2": types.StringValue("value2"),
	})

	mapDifferentKeys := NewTypedMap(map[string]types.String{
		"key1": types.StringValue("value1"),
		"key3": types.StringValue("value3"),
	})

	mapDifferentLength := NewTypedMap(map[string]types.String{
		"key1": types.StringValue("value1"),
		"key2": types.StringValue("value2"),
		"key3": types.StringValue("value3"),
	})

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
			map1:     NewTypedMapNull[types.String](),
			map2:     map1,
			expected: false,
		},
		"unknown != known": {
			map1:     NewTypedMapUnknown[types.String](),
			map2:     map1,
			expected: false,
		},
		"null == null": {
			map1:     NewTypedMapNull[types.String](),
			map2:     NewTypedMapNull[types.String](),
			expected: true,
		},
		"unknown == unknown": {
			map1:     NewTypedMapUnknown[types.String](),
			map2:     NewTypedMapUnknown[types.String](),
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

func TestTypedMapHas(t *testing.T) {
	t.Parallel()

	typedMap := NewTypedMap(map[string]types.String{
		"key1": types.StringValue("value1"),
		"key2": types.StringValue("value2"),
	})

	emptyMap := NewTypedMap(map[string]types.String{})

	nullMap := NewTypedMapNull[types.String]()

	unknownMap := NewTypedMapUnknown[types.String]()

	testcases := map[string]struct {
		typedMap TypedMap[types.String]
		key      string
		expected bool
	}{
		"existing key": {
			typedMap: typedMap,
			key:      "key1",
			expected: true,
		},
		"another existing key": {
			typedMap: typedMap,
			key:      "key2",
			expected: true,
		},
		"non-existing key": {
			typedMap: typedMap,
			key:      "key3",
			expected: false,
		},
		"empty string key": {
			typedMap: typedMap,
			key:      "",
			expected: false,
		},
		"empty map": {
			typedMap: emptyMap,
			key:      "key1",
			expected: false,
		},
		"null map": {
			typedMap: nullMap,
			key:      "key1",
			expected: false,
		},
		"unknown map": {
			typedMap: unknownMap,
			key:      "key1",
			expected: false,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testcase.expected, testcase.typedMap.Has(testcase.key))
		})
	}
}

func TestTypedMapSet(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		initialMap TypedMap[types.String]
		key        string
		value      types.String
		verify     func(t *testing.T, m TypedMap[types.String])
	}{
		"set new key": {
			initialMap: NewTypedMap(map[string]types.String{
				"key1": types.StringValue("value1"),
			}),
			key:   "key2",
			value: types.StringValue("value2"),
			verify: func(t *testing.T, m TypedMap[types.String]) {
				t.Helper()
				assert.True(t, m.Has("key1"))
				assert.True(t, m.Has("key2"))
				assert.Equal(t, types.StringValue("value1"), m.Elements()["key1"])
				assert.Equal(t, types.StringValue("value2"), m.Elements()["key2"])
			},
		},
		"overwrite existing key": {
			initialMap: NewTypedMap(map[string]types.String{
				"key1": types.StringValue("value1"),
			}),
			key:   "key1",
			value: types.StringValue("newvalue"),
			verify: func(t *testing.T, m TypedMap[types.String]) {
				t.Helper()
				assert.True(t, m.Has("key1"))
				assert.Equal(t, types.StringValue("newvalue"), m.Elements()["key1"])
			},
		},
		"set in empty map": {
			initialMap: NewTypedMap(map[string]types.String{}),
			key:        "key1",
			value:      types.StringValue("value1"),
			verify: func(t *testing.T, m TypedMap[types.String]) {
				t.Helper()
				assert.True(t, m.Has("key1"))
				assert.Equal(t, types.StringValue("value1"), m.Elements()["key1"])
			},
		},
		"set with empty string key": {
			initialMap: NewTypedMap(map[string]types.String{}),
			key:        "",
			value:      types.StringValue("value"),
			verify: func(t *testing.T, m TypedMap[types.String]) {
				t.Helper()
				assert.True(t, m.Has(""))
				assert.Equal(t, types.StringValue("value"), m.Elements()[""])
			},
		},
		"set null value": {
			initialMap: NewTypedMap(map[string]types.String{}),
			key:        "key1",
			value:      types.StringNull(),
			verify: func(t *testing.T, m TypedMap[types.String]) {
				t.Helper()
				assert.True(t, m.Has("key1"))
				assert.Equal(t, types.StringNull(), m.Elements()["key1"])
			},
		},
		"set unknown value": {
			initialMap: NewTypedMap(map[string]types.String{}),
			key:        "key1",
			value:      types.StringUnknown(),
			verify: func(t *testing.T, m TypedMap[types.String]) {
				t.Helper()
				assert.True(t, m.Has("key1"))
				assert.Equal(t, types.StringUnknown(), m.Elements()["key1"])
			},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			testcase.initialMap.Set(testcase.key, testcase.value)
			testcase.verify(t, testcase.initialMap)
		})
	}
}
