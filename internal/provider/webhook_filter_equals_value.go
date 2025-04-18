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

type WebhookFilterEqualsValue struct {
	Doc   types.String `tfsdk:"doc"`
	Value types.String `tfsdk:"value"`
	state attr.ValueState
}

func NewWebhookFilterEqualsValueKnown() WebhookFilterEqualsValue {
	return WebhookFilterEqualsValue{
		Doc:   types.StringNull(),
		Value: types.StringNull(),
		state: attr.ValueStateKnown,
	}
}

func NewWebhookFilterEqualsValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (WebhookFilterEqualsValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := WebhookFilterEqualsValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
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
	return WebhookFilterEqualsType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = WebhookFilterEqualsValue{}

//nolint:ireturn
func (v WebhookFilterEqualsValue) Type(ctx context.Context) attr.Type {
	return WebhookFilterEqualsType{ObjectType: v.ObjectType(ctx)}
}

func (v WebhookFilterEqualsValue) ObjectType(ctx context.Context) types.ObjectType {
	return types.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v WebhookFilterEqualsValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v WebhookFilterEqualsValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[WebhookFilterEqualsValue](v, o)
}

func (v WebhookFilterEqualsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v WebhookFilterEqualsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v WebhookFilterEqualsValue) String() string {
	return "WebhookFilterEqualsValue"
}

func (v WebhookFilterEqualsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v WebhookFilterEqualsValue) ToObjectValue(ctx context.Context) (types.Object, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
