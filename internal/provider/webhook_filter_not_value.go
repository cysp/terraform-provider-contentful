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

type WebhookFilterNotValue struct {
	Equals WebhookFilterEqualsValue `tfsdk:"equals"`
	In     WebhookFilterInValue     `tfsdk:"in"`
	Regexp WebhookFilterRegexpValue `tfsdk:"regexp"`
	state  attr.ValueState
}

func NewWebhookFilterNotValueKnown() WebhookFilterNotValue {
	return WebhookFilterNotValue{
		Equals: NewWebhookFilterEqualsValueNull(),
		In:     NewWebhookFilterInValueNull(),
		Regexp: NewWebhookFilterRegexpValueNull(),
		state:  attr.ValueStateKnown,
	}
}

func NewWebhookFilterNotValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (WebhookFilterNotValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := WebhookFilterNotValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewWebhookFilterNotValueNull() WebhookFilterNotValue {
	return WebhookFilterNotValue{
		state: attr.ValueStateNull,
	}
}

func NewWebhookFilterNotValueUnknown() WebhookFilterNotValue {
	return WebhookFilterNotValue{
		state: attr.ValueStateUnknown,
	}
}

func (v WebhookFilterNotValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"equals": schema.SingleNestedAttribute{
			Attributes: WebhookFilterEqualsValue{}.SchemaAttributes(ctx),
			CustomType: WebhookFilterEqualsValue{}.CustomType(ctx),
			Optional:   true,
		},
		"in": schema.SingleNestedAttribute{
			Attributes: WebhookFilterInValue{}.SchemaAttributes(ctx),
			CustomType: WebhookFilterInValue{}.CustomType(ctx),
			Optional:   true,
		},
		"regexp": schema.SingleNestedAttribute{
			Attributes: WebhookFilterRegexpValue{}.SchemaAttributes(ctx),
			CustomType: WebhookFilterRegexpValue{}.CustomType(ctx),
			Optional:   true,
		},
	}
}

//nolint:ireturn
func (v WebhookFilterNotValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return WebhookFilterNotType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = WebhookFilterNotValue{}

//nolint:ireturn
func (v WebhookFilterNotValue) Type(ctx context.Context) attr.Type {
	return WebhookFilterNotType{ObjectType: v.ObjectType(ctx)}
}

func (v WebhookFilterNotValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v WebhookFilterNotValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v WebhookFilterNotValue) Equal(o attr.Value) bool {
	other, ok := o.(WebhookFilterNotValue)
	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state == attr.ValueStateKnown {
		return tpfr.ValueAttributesEqual(v, other)
	}

	return true
}

func (v WebhookFilterNotValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v WebhookFilterNotValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v WebhookFilterNotValue) String() string {
	return ""
}

func (v WebhookFilterNotValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v WebhookFilterNotValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
