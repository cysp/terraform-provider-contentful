package tf

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func NewEmptySetMust(elementType attr.Type) basetypes.SetValue {
	list, _ := types.SetValue(elementType, []attr.Value{})

	return list
}
