package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type WebhookFilterInValue struct {
	Doc    types.String `tfsdk:"doc"`
	Values types.List   `tfsdk:"values"`
	state  attr.ValueState
}

func NewWebhookFilterInValueKnown() WebhookFilterInValue {
	return WebhookFilterInValue{
		Doc:    types.StringNull(),
		Values: types.ListNull(types.StringType),
		state:  attr.ValueStateKnown,
	}
}

func NewWebhookFilterInValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (WebhookFilterInValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := WebhookFilterInValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := setTFSDKAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewWebhookFilterInValueNull() WebhookFilterInValue {
	return WebhookFilterInValue{
		state: attr.ValueStateNull,
	}
}

func NewWebhookFilterInValueUnknown() WebhookFilterInValue {
	return WebhookFilterInValue{
		state: attr.ValueStateUnknown,
	}
}

func (v WebhookFilterInValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"doc": schema.StringAttribute{
			Required: true,
		},
		"values": schema.ListAttribute{
			ElementType: basetypes.StringType{},
			Required:    true,
		},
	}
}

//nolint:ireturn
func (v WebhookFilterInValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return WebhookFilterInType{
		v.ObjectType(ctx),
	}
}

var _ basetypes.ObjectValuable = WebhookFilterInValue{}

//nolint:ireturn
func (v WebhookFilterInValue) Type(ctx context.Context) attr.Type {
	return WebhookFilterInType{
		ObjectType: v.ObjectType(ctx),
	}
}

func (v WebhookFilterInValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: v.ObjectAttrTypes(ctx),
	}
}

func (v WebhookFilterInValue) ObjectAttrTypes(_ context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"doc": types.StringType,
		"values": types.ListType{
			ElemType: types.StringType,
		},
	}
}

func (v WebhookFilterInValue) Equal(o attr.Value) bool {
	other, ok := o.(WebhookFilterInValue)
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

func (v WebhookFilterInValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v WebhookFilterInValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v WebhookFilterInValue) String() string {
	panic("unimplemented")
}

//nolint:dupl
func (v WebhookFilterInValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	tft := WebhookFilterInType{}.TerraformType(ctx)

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
	val := make(map[string]tftypes.Value, 2)

	var docErr error
	val["doc"], docErr = v.Doc.ToTerraformValue(ctx)

	var valuesErr error
	val["values"], valuesErr = v.Values.ToTerraformValue(ctx)

	validateErr := tftypes.ValidateValue(tft, val)

	err := errors.Join(docErr, valuesErr, validateErr)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(tft, val), nil
}

func (v WebhookFilterInValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	attributeTypes := v.ObjectAttrTypes(ctx)

	switch {
	case v.IsNull():
		return types.ObjectNull(attributeTypes), nil
	case v.IsUnknown():
		return types.ObjectUnknown(attributeTypes), nil
	}

	attributes := map[string]attr.Value{
		"doc":    v.Doc,
		"values": v.Values,
	}

	return types.ObjectValue(attributeTypes, attributes)
}
