package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type WebhookFilterRegexpValue struct {
	Doc     types.String `tfsdk:"doc"`
	Pattern types.String `tfsdk:"pattern"`
	state   attr.ValueState
}

func NewWebhookFilterRegexpValueKnown() WebhookFilterRegexpValue {
	return WebhookFilterRegexpValue{
		Doc:     types.StringNull(),
		Pattern: types.StringNull(),
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

func (v WebhookFilterRegexpValue) ObjectType(ctx context.Context) types.ObjectType {
	return types.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v WebhookFilterRegexpValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v WebhookFilterRegexpValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[WebhookFilterRegexpValue](v, o)
}

func (v WebhookFilterRegexpValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v WebhookFilterRegexpValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v WebhookFilterRegexpValue) String() string {
	return "WebhookFilterRegexpValue"
}

func (v WebhookFilterRegexpValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v WebhookFilterRegexpValue) ToObjectValue(ctx context.Context) (types.Object, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
