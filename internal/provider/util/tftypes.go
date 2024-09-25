package util

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func NewEmptyListMust(elementType attr.Type) basetypes.ListValue {
	list, _ := types.ListValue(elementType, []attr.Value{})

	return list
}

func NewEmptySetMust(elementType attr.Type) basetypes.SetValue {
	list, _ := types.SetValue(elementType, []attr.Value{})

	return list
}
