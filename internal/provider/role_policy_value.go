package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type RolePolicyValue struct {
	Actions    TypedList[types.String] `tfsdk:"actions"`
	Constraint jsontypes.Normalized    `tfsdk:"constraint"`
	Effect     types.String            `tfsdk:"effect"`
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

//nolint:ireturn
func (v RolePolicyValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return RolePolicyType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = RolePolicyValue{}

//nolint:ireturn
func (v RolePolicyValue) Type(ctx context.Context) attr.Type {
	return RolePolicyType{ObjectType: v.ObjectType(ctx)}
}

func (v RolePolicyValue) ObjectType(ctx context.Context) types.ObjectType {
	return types.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v RolePolicyValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v RolePolicyValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[RolePolicyValue](v, o)
}

func (v RolePolicyValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v RolePolicyValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v RolePolicyValue) String() string {
	return "RolePolicyValue"
}

func (v RolePolicyValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v RolePolicyValue) ToObjectValue(ctx context.Context) (types.Object, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
