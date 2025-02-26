package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

func NewWebhookTransformationValueKnownFromAttributes(_ context.Context, attributes map[string]attr.Value) (WebhookTransformationValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	methodValue, methodOk := attributes["method"].(types.String)
	if !methodOk {
		diags.AddAttributeError(path.Root("method"), "invalid data", fmt.Sprintf("expected string, got %T", attributes["method"]))
	}

	contentTypeValue, contentTypeOk := attributes["content_type"].(types.String)
	if !contentTypeOk {
		diags.AddAttributeError(path.Root("content_type"), "invalid data", fmt.Sprintf("expected string, got %T", attributes["content_type"]))
	}

	includeContentLengthValue, includeContentLengthOk := attributes["include_content_length"].(types.Bool)
	if !includeContentLengthOk {
		diags.AddAttributeError(path.Root("include_content_length"), "invalid data", fmt.Sprintf("expected bool, got %T", attributes["include_content_length"]))
	}

	bodyValue, contentTypeOk := attributes["body"].(jsontypes.Normalized)
	if !contentTypeOk {
		diags.AddAttributeError(path.Root("body"), "invalid data", fmt.Sprintf("expected json string, got %T", attributes["body"]))
	}

	return WebhookTransformationValue{
		Method:               methodValue,
		ContentType:          contentTypeValue,
		IncludeContentLength: includeContentLengthValue,
		Body:                 bodyValue,
		state:                attr.ValueStateKnown,
	}, diags
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
	return WebhookTransformationType{
		v.ObjectType(ctx),
	}
}

var _ basetypes.ObjectValuable = WebhookTransformationValue{}

//nolint:ireturn
func (v WebhookTransformationValue) Type(ctx context.Context) attr.Type {
	return WebhookTransformationType{
		ObjectType: v.ObjectType(ctx),
	}
}

func (v WebhookTransformationValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: v.ObjectAttrTypes(ctx),
	}
}

func (v WebhookTransformationValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ReflectObjectAttrTypes(ctx, v)
}

func (v WebhookTransformationValue) Equal(o attr.Value) bool {
	return ReflectAttrValueEqual(v, o)
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

type TerraformValuable interface {
	ToTerraformValue(_ context.Context) (tftypes.Value, error)
}

func (v WebhookTransformationValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	return ReflectIntoTerraformValue(ctx, v)
}

func (v WebhookTransformationValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return ReflectIntoObjectValue(ctx, v)
}
