package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type RolePolicyValue struct {
	Actions    types.List           `tfsdk:"actions"`
	Constraint jsontypes.Normalized `tfsdk:"constraint"`
	Effect     types.String         `tfsdk:"effect"`
	state      attr.ValueState
}

func NewRolePolicyValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (RolePolicyValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := RolePolicyValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewRolePolicyValueNull() RolePolicyValue {
	return RolePolicyValue{
		state: attr.ValueStateNull,
	}
}

func NewRolePolicyValueUnknown() RolePolicyValue {
	return RolePolicyValue{
		state: attr.ValueStateUnknown,
	}
}

func (v RolePolicyValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"actions": schema.ListAttribute{
			ElementType: types.StringType,
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

//nolint:ireturn
func (v RolePolicyValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return RolePolicyType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = RolePolicyValue{}

//nolint:ireturn
func (v RolePolicyValue) Type(ctx context.Context) attr.Type {
	return RolePolicyType{ObjectType: v.ObjectType(ctx)}
}

func (v RolePolicyValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v RolePolicyValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v RolePolicyValue) Equal(o attr.Value) bool {
	other, ok := o.(RolePolicyValue)
	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state == attr.ValueStateKnown {
		return compareTFSDKAttributesEqual(v, other)
	}

	return true
}

func (v RolePolicyValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v RolePolicyValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v RolePolicyValue) String() string {
	return ""
}

func (v RolePolicyValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	return ReflectToTerraformValue(ctx, v, v.state)
}

func (v RolePolicyValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
