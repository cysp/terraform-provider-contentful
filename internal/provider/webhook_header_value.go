package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

//nolint:recvcheck
type WebhookHeaderValue struct {
	Value  basetypes.StringValue `tfsdk:"value"`
	Secret basetypes.BoolValue   `tfsdk:"secret"`
	state  attr.ValueState
}

func NewWebhookHeaderValueKnown() WebhookHeaderValue {
	return WebhookHeaderValue{
		state: attr.ValueStateKnown,
	}
}

func NewWebhookHeaderValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (WebhookHeaderValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := WebhookHeaderValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := setTFSDKAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewWebhookHeaderValueKnownFromAttributesMust(ctx context.Context, attributes map[string]attr.Value) WebhookHeaderValue {
	value, diags := NewWebhookHeaderValueKnownFromAttributes(ctx, attributes)
	if diags.HasError() {
		panic(diags)
	}

	return value
}

func NewWebhookHeaderValueNull() WebhookHeaderValue {
	return WebhookHeaderValue{
		state: attr.ValueStateNull,
	}
}

func NewWebhookHeaderValueUnknown() WebhookHeaderValue {
	return WebhookHeaderValue{
		state: attr.ValueStateUnknown,
	}
}

func (v WebhookHeaderValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"value": schema.StringAttribute{
			Required: true,
		},
		"secret": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

//nolint:ireturn
func (v WebhookHeaderValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return WebhookHeaderType{
		v.ObjectType(ctx),
	}
}

var _ basetypes.ObjectValuable = WebhookHeaderValue{}

//nolint:ireturn
func (v WebhookHeaderValue) Type(ctx context.Context) attr.Type {
	return WebhookHeaderType{
		ObjectType: v.ObjectType(ctx),
	}
}

func (v WebhookHeaderValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: v.ObjectAttrTypes(ctx),
	}
}

func (v WebhookHeaderValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"value":  types.String{}.Type(ctx),
		"secret": types.Bool{}.Type(ctx),
	}
}

func (v WebhookHeaderValue) Equal(o attr.Value) bool {
	other, ok := o.(WebhookHeaderValue)
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

func (v WebhookHeaderValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v WebhookHeaderValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v WebhookHeaderValue) String() string {
	return "WebhookHeaderValue"
}

//nolint:dupl
func (v WebhookHeaderValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	tft := WebhookHeaderType{}.TerraformType(ctx)

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
	val := make(map[string]tftypes.Value, 4)

	var valueErr error
	val["value"], valueErr = v.Value.ToTerraformValue(ctx)

	var secretErr error
	val["secret"], secretErr = v.Secret.ToTerraformValue(ctx)

	validateErr := tftypes.ValidateValue(tft, val)

	err := errors.Join(valueErr, secretErr, validateErr)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(tft, val), nil
}

func (v WebhookHeaderValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	attributeTypes := v.ObjectAttrTypes(ctx)

	switch {
	case v.IsNull():
		return types.ObjectNull(attributeTypes), nil
	case v.IsUnknown():
		return types.ObjectUnknown(attributeTypes), nil
	}

	attributes := map[string]attr.Value{
		"value":  v.Value,
		"secret": v.Secret,
	}

	return types.ObjectValue(attributeTypes, attributes)
}
