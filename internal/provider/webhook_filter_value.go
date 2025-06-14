package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type WebhookFilterValue struct {
	Not    WebhookFilterNotValue    `tfsdk:"not"`
	Equals WebhookFilterEqualsValue `tfsdk:"equals"`
	In     WebhookFilterInValue     `tfsdk:"in"`
	Regexp WebhookFilterRegexpValue `tfsdk:"regexp"`
	state  attr.ValueState
}

func NewWebhookFilterValueKnown() WebhookFilterValue {
	return WebhookFilterValue{
		Not:    NewWebhookFilterNotValueNull(),
		Equals: NewWebhookFilterEqualsValueNull(),
		In:     NewWebhookFilterInValueNull(),
		Regexp: NewWebhookFilterRegexpValueNull(),
		state:  attr.ValueStateKnown,
	}
}

func NewWebhookFilterValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (WebhookFilterValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := WebhookFilterValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewWebhookFilterValueNull() WebhookFilterValue {
	return WebhookFilterValue{
		state: attr.ValueStateNull,
	}
}

func NewWebhookFilterValueUnknown() WebhookFilterValue {
	return WebhookFilterValue{
		state: attr.ValueStateUnknown,
	}
}

//nolint:ireturn
func (v WebhookFilterValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return WebhookFilterType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = WebhookFilterValue{}

//nolint:ireturn
func (v WebhookFilterValue) Type(ctx context.Context) attr.Type {
	return WebhookFilterType{ObjectType: v.ObjectType(ctx)}
}

func (v WebhookFilterValue) ObjectType(ctx context.Context) types.ObjectType {
	return types.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v WebhookFilterValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v WebhookFilterValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[WebhookFilterValue](v, o)
}

func (v WebhookFilterValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v WebhookFilterValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v WebhookFilterValue) String() string {
	return "WebhookFilterValue"
}

func (v WebhookFilterValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v WebhookFilterValue) ToObjectValue(ctx context.Context) (types.Object, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
