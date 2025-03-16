package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type WebhookFilterRegexpValue struct {
	Doc     basetypes.StringValue `tfsdk:"doc"`
	Pattern basetypes.StringValue `tfsdk:"pattern"`
	state   attr.ValueState
}

func NewWebhookFilterRegexpValueKnown() WebhookFilterRegexpValue {
	return WebhookFilterRegexpValue{
		Doc:     basetypes.NewStringNull(),
		Pattern: basetypes.NewStringNull(),
		state:   attr.ValueStateKnown,
	}
}

func NewWebhookFilterRegexpValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (WebhookFilterRegexpValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := WebhookFilterRegexpValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewWebhookFilterRegexpValueNull() WebhookFilterRegexpValue {
	return WebhookFilterRegexpValue{
		state: attr.ValueStateNull,
	}
}

func NewWebhookFilterRegexpValueUnknown() WebhookFilterRegexpValue {
	return WebhookFilterRegexpValue{
		state: attr.ValueStateUnknown,
	}
}

func (v WebhookFilterRegexpValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"doc": schema.StringAttribute{
			Required: true,
		},
		"pattern": schema.StringAttribute{
			Required: true,
		},
	}
}

//nolint:ireturn
func (v WebhookFilterRegexpValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return WebhookFilterRegexpType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = WebhookFilterRegexpValue{}

//nolint:ireturn
func (v WebhookFilterRegexpValue) Type(ctx context.Context) attr.Type {
	return WebhookFilterRegexpType{ObjectType: v.ObjectType(ctx)}
}

func (v WebhookFilterRegexpValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v WebhookFilterRegexpValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v WebhookFilterRegexpValue) Equal(o attr.Value) bool {
	other, ok := o.(WebhookFilterRegexpValue)
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

func (v WebhookFilterRegexpValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v WebhookFilterRegexpValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v WebhookFilterRegexpValue) String() string {
	panic("unimplemented")
}

func (v WebhookFilterRegexpValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	return ReflectToTerraformValue(ctx, v, v.state)
}

func (v WebhookFilterRegexpValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
