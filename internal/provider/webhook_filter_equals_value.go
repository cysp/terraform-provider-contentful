package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type WebhookFilterEqualsValue struct {
	Doc   basetypes.StringValue `tfsdk:"doc"`
	Value basetypes.StringValue `tfsdk:"value"`
	state attr.ValueState
}

func NewWebhookFilterEqualsValueKnown() WebhookFilterEqualsValue {
	return WebhookFilterEqualsValue{
		Doc:   basetypes.NewStringNull(),
		Value: basetypes.NewStringNull(),
		state: attr.ValueStateKnown,
	}
}

func NewWebhookFilterEqualsValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (WebhookFilterEqualsValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := WebhookFilterEqualsValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := setTFSDKAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewWebhookFilterEqualsValueNull() WebhookFilterEqualsValue {
	return WebhookFilterEqualsValue{
		state: attr.ValueStateNull,
	}
}

func NewWebhookFilterEqualsValueUnknown() WebhookFilterEqualsValue {
	return WebhookFilterEqualsValue{
		state: attr.ValueStateUnknown,
	}
}

func (v WebhookFilterEqualsValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"doc": schema.StringAttribute{
			Required: true,
		},
		"value": schema.StringAttribute{
			Required: true,
		},
	}
}

//nolint:ireturn
func (v WebhookFilterEqualsValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return WebhookFilterEqualsType{
		v.ObjectType(ctx),
	}
}

var _ basetypes.ObjectValuable = WebhookFilterEqualsValue{}

//nolint:ireturn
func (v WebhookFilterEqualsValue) Type(ctx context.Context) attr.Type {
	return WebhookFilterEqualsType{
		ObjectType: v.ObjectType(ctx),
	}
}

func (v WebhookFilterEqualsValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: v.ObjectAttrTypes(ctx),
	}
}

func (v WebhookFilterEqualsValue) ObjectAttrTypes(_ context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"doc":   types.StringType,
		"value": types.StringType,
	}
}

func (v WebhookFilterEqualsValue) Equal(o attr.Value) bool {
	other, ok := o.(WebhookFilterEqualsValue)
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

func (v WebhookFilterEqualsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v WebhookFilterEqualsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v WebhookFilterEqualsValue) String() string {
	return ""
}

func (v WebhookFilterEqualsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	return ReflectToTerraformValue(ctx, v, v.state)
}

func (v WebhookFilterEqualsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return ReflectToObjectValue(ctx, v)
}
