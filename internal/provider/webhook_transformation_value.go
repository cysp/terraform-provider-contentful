package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type WebhookTransformationValue struct {
	Method               basetypes.StringValue `tfsdk:"method"`
	ContentType          basetypes.StringValue `tfsdk:"content_type"`
	IncludeContentLength basetypes.BoolValue   `tfsdk:"include_content_length"`
	Body                 jsontypes.Normalized  `tfsdk:"body"`
	state                attr.ValueState
}

func NewWebhookTransformationValueKnown() WebhookTransformationValue {
	return WebhookTransformationValue{
		state: attr.ValueStateKnown,
	}
}

func NewWebhookTransformationValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (WebhookTransformationValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := WebhookTransformationValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := setTFSDKAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewWebhookTransformationValueKnownFromAttributesMust(ctx context.Context, attributes map[string]attr.Value) WebhookTransformationValue {
	value, diags := NewWebhookTransformationValueKnownFromAttributes(ctx, attributes)
	if diags.HasError() {
		panic(diags)
	}

	return value
}

func NewWebhookTransformationValueNull() WebhookTransformationValue {
	return WebhookTransformationValue{
		state: attr.ValueStateNull,
	}
}

func NewWebhookTransformationValueUnknown() WebhookTransformationValue {
	return WebhookTransformationValue{
		state: attr.ValueStateUnknown,
	}
}

func (v WebhookTransformationValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"method": schema.StringAttribute{
			Optional: true,
		},
		"content_type": schema.StringAttribute{
			Optional: true,
		},
		"include_content_length": schema.BoolAttribute{
			Optional: true,
		},
		"body": schema.StringAttribute{
			Optional:   true,
			CustomType: jsontypes.NormalizedType{},
		},
	}
}

//nolint:ireturn
func (v WebhookTransformationValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return WebhookTransformationType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = WebhookTransformationValue{}

//nolint:ireturn
func (v WebhookTransformationValue) Type(ctx context.Context) attr.Type {
	return WebhookTransformationType{ObjectType: v.ObjectType(ctx)}
}

func (v WebhookTransformationValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v WebhookTransformationValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"method":                 types.String{}.Type(ctx),
		"content_type":           types.String{}.Type(ctx),
		"include_content_length": types.Bool{}.Type(ctx),
		"body":                   jsontypes.Normalized{}.Type(ctx),
	}
}

func (v WebhookTransformationValue) Equal(o attr.Value) bool {
	other, ok := o.(WebhookTransformationValue)
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

func (v WebhookTransformationValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v WebhookTransformationValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v WebhookTransformationValue) String() string {
	return "WebhookTransformationValue"
}

func (v WebhookTransformationValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	return ReflectToTerraformValue(ctx, v, v.state)
}

func (v WebhookTransformationValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return ReflectToObjectValue(ctx, v)
}
