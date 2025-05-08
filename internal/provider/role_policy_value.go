package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RolePolicyValue struct {
	Actions    []types.String       `tfsdk:"actions"`
	Constraint jsontypes.Normalized `tfsdk:"constraint"`
	Effect     types.String         `tfsdk:"effect"`
}

func (v RolePolicyValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"actions": schema.ListAttribute{
			ElementType: types.StringType,
			CustomType:  TypedList[types.String]{}.CustomType(ctx),
			Required:    true,
		},
		"constraint": schema.StringAttribute{
			CustomType: jsontypes.NormalizedType{},
			Optional:   true,
		},
		"effect": schema.StringAttribute{
			Required: true,
		},
	}
}
