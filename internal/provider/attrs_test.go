package provider_test

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type AttrValueWithToObjectValue interface {
	attr.Value

	ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics)
}
