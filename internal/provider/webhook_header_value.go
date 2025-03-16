package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
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
	return WebhookHeaderType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = WebhookHeaderValue{}

//nolint:ireturn
func (v WebhookHeaderValue) Type(ctx context.Context) attr.Type {
	return WebhookHeaderType{ObjectType: v.ObjectType(ctx)}
}

func (v WebhookHeaderValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v WebhookHeaderValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
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

func (v WebhookHeaderValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	return ReflectToTerraformValue(ctx, v, v.state)
}

func (v WebhookHeaderValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
