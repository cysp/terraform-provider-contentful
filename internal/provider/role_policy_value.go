package provider

import (
	"context"
	"errors"
	"fmt"

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

	setAttributesDiags := setTFSDKAttributesInValue(ctx, &value, attributes)
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
	return RolePolicyType{
		v.ObjectType(ctx),
	}
}

var _ basetypes.ObjectValuable = RolePolicyValue{}

//nolint:ireturn
func (v RolePolicyValue) Type(ctx context.Context) attr.Type {
	return RolePolicyType{
		ObjectType: v.ObjectType(ctx),
	}
}

func (v RolePolicyValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: v.ObjectAttrTypes(ctx),
	}
}

func (v RolePolicyValue) ObjectAttrTypes(_ context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"actions":    types.ListType{ElemType: types.StringType},
		"constraint": jsontypes.NormalizedType{},
		"effect":     types.StringType,
	}
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

//nolint:dupl
func (v RolePolicyValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	tft := RolePolicyType{}.TerraformType(ctx)

	switch v.state {
	case attr.ValueStateKnown:
		break
	case attr.ValueStateNull:
		return tftypes.NewValue(tft, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(tft, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}

	//nolint:gomnd,mnd
	val := make(map[string]tftypes.Value, 3)

	var actionsErr error
	val["actions"], actionsErr = v.Actions.ToTerraformValue(ctx)

	var constraintErr error
	val["constraint"], constraintErr = v.Constraint.ToTerraformValue(ctx)

	var effectErr error
	val["effect"], effectErr = v.Effect.ToTerraformValue(ctx)

	validateErr := tftypes.ValidateValue(tft, val)

	err := errors.Join(actionsErr, constraintErr, effectErr, validateErr)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(tft, val), nil
}

func (v RolePolicyValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	attributeTypes := v.ObjectAttrTypes(ctx)

	switch {
	case v.IsNull():
		return types.ObjectNull(attributeTypes), nil
	case v.IsUnknown():
		return types.ObjectUnknown(attributeTypes), nil
	}

	attributes := map[string]attr.Value{
		"actions":    v.Actions,
		"constraint": v.Constraint,
		"effect":     v.Effect,
	}

	return types.ObjectValue(attributeTypes, attributes)
}
