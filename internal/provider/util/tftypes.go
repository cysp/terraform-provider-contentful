package util

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func NewEmptyListMust(elementType attr.Type) basetypes.ListValue {
	list, _ := types.ListValue(elementType, []attr.Value{})

	return list
}

func NewEmptySetMust(elementType attr.Type) basetypes.SetValue {
	list, _ := types.SetValue(elementType, []attr.Value{})

	return list
}

func AttributesFromTerraform(ctx context.Context, attrTypes map[string]attr.Type, value tftypes.Value) (map[string]attr.Value, error) {
	attributes := map[string]attr.Value{}

	tfvals := map[string]tftypes.Value{}

	err := value.As(&tfvals)
	if err != nil {
		return nil, err
	}

	for key, tfval := range tfvals {
		a, err := attrTypes[key].ValueFromTerraform(ctx, tfval)
		if err != nil {
			return nil, err
		}

		attributes[key] = a
	}

	return attributes, nil
}
