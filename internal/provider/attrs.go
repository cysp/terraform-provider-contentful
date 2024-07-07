package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type AttrTypeWithValueFromObject interface {
	attr.Type

	ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics)
}

type AttrValueWithToObjectValue interface {
	attr.Value

	ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics)
}
