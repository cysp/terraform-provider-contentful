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

type WebhookFilterInValue struct {
	Doc    types.String `tfsdk:"doc"`
	Values types.List   `tfsdk:"values"`
	state  attr.ValueState
}

func NewWebhookFilterInValueKnown() WebhookFilterInValue {
	return WebhookFilterInValue{
		Doc:    types.StringNull(),
		Values: types.ListNull(types.StringType),
		state:  attr.ValueStateKnown,
	}
}

func NewWebhookFilterInValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (WebhookFilterInValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := WebhookFilterInValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := setTFSDKAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewWebhookFilterInValueNull() WebhookFilterInValue {
	return WebhookFilterInValue{
		state: attr.ValueStateNull,
	}
}

func NewWebhookFilterInValueUnknown() WebhookFilterInValue {
	return WebhookFilterInValue{
		state: attr.ValueStateUnknown,
	}
}

func (v WebhookFilterInValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"doc": schema.StringAttribute{
			Required: true,
		},
		"values": schema.ListAttribute{
			ElementType: basetypes.StringType{},
			Required:    true,
		},
	}
}

//nolint:ireturn
func (v WebhookFilterInValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return WebhookFilterInType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = WebhookFilterInValue{}

//nolint:ireturn
func (v WebhookFilterInValue) Type(ctx context.Context) attr.Type {
	return WebhookFilterInType{ObjectType: v.ObjectType(ctx)}
}

func (v WebhookFilterInValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v WebhookFilterInValue) ObjectAttrTypes(_ context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"doc": types.StringType,
		"values": types.ListType{
			ElemType: types.StringType,
		},
	}
}

func (v WebhookFilterInValue) Equal(o attr.Value) bool {
	other, ok := o.(WebhookFilterInValue)
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

func (v WebhookFilterInValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v WebhookFilterInValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v WebhookFilterInValue) String() string {
	panic("unimplemented")
}

func (v WebhookFilterInValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	return ReflectToTerraformValue(ctx, v, v.state)
}

func (v WebhookFilterInValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return ReflectToObjectValue(ctx, v)
}
