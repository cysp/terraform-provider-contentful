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
		return v.Method.Equal(other.Method) && v.ContentType.Equal(other.ContentType) && v.IncludeContentLength.Equal(other.IncludeContentLength) && v.Body.Equal(other.Body)
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

//nolint:dupl
func (v WebhookTransformationValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	tft := WebhookTransformationType{}.TerraformType(ctx)

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

	var methodErr error
	val["method"], methodErr = v.Method.ToTerraformValue(ctx)

	var contentTypeErr error
	val["content_type"], contentTypeErr = v.ContentType.ToTerraformValue(ctx)

	var includeContentLengthErr error
	val["include_content_length"], includeContentLengthErr = v.IncludeContentLength.ToTerraformValue(ctx)

	var bodyErr error
	val["body"], bodyErr = v.Body.ToTerraformValue(ctx)

	validateErr := tftypes.ValidateValue(tft, val)

	err := errors.Join(methodErr, contentTypeErr, includeContentLengthErr, bodyErr, validateErr)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(tft, val), nil
}

func (v WebhookTransformationValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	attributeTypes := v.ObjectAttrTypes(ctx)

	switch {
	case v.IsNull():
		return types.ObjectNull(attributeTypes), nil
	case v.IsUnknown():
		return types.ObjectUnknown(attributeTypes), nil
	}

	attributes := map[string]attr.Value{
		"method":                 v.Method,
		"content_type":           v.ContentType,
		"include_content_length": v.IncludeContentLength,
		"body":                   v.Body,
	}

	return types.ObjectValue(attributeTypes, attributes)
}
