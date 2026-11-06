package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestTypedCollectionTypeEqual(t *testing.T) {
	t.Parallel()

	listType := TypedList[types.String]{}.CustomType(t.Context())
	mapType := TypedMap[types.String]{}.CustomType(t.Context())

	for name, test := range map[string]struct {
		left  attr.Type
		right attr.Type
		equal bool
	}{
		"same list type":                    {left: listType, right: TypedList[types.String]{}.CustomType(t.Context()), equal: true},
		"same map type":                     {left: mapType, right: TypedMap[types.String]{}.CustomType(t.Context()), equal: true},
		"different collections":             {left: listType, right: mapType},
		"different list element":            {left: listType, right: TypedList[types.Bool]{}.CustomType(t.Context())},
		"different map element":             {left: mapType, right: TypedMap[types.Bool]{}.CustomType(t.Context())},
		"different list Go type":            {left: listType, right: types.ListType{ElemType: types.StringType}},
		"different map Go type":             {left: mapType, right: types.MapType{ElemType: types.StringType}},
		"different overridden list element": {left: listType.WithElementType(types.StringType), right: listType.WithElementType(types.BoolType)},
		"different overridden map element":  {left: mapType.WithElementType(types.StringType), right: mapType.WithElementType(types.BoolType)},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.equal, test.left.Equal(test.right))
			assert.Equal(t, test.equal, test.right.Equal(test.left))
		})
	}
}
