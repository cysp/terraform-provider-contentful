package webhookfilter

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type WebhookFilterInValue struct {
	Doc    basetypes.StringValue `tfsdk:"doc"`
	Values basetypes.ListValue   `tfsdk:"values"`
	state  attr.ValueState
}

func NewWebhookFilterInValueKnown() WebhookFilterInValue {
	return WebhookFilterInValue{
		state: attr.ValueStateKnown,
	}
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

// Equal implements basetypes.ObjectValuable.
func (m WebhookFilterInValue) Equal(attr.Value) bool {
	//xxx
	return true
}

// IsNull implements basetypes.ObjectValuable.
func (m WebhookFilterInValue) IsNull() bool {
	return m.state == attr.ValueStateNull
}

// IsUnknown implements basetypes.ObjectValuable.
func (m WebhookFilterInValue) IsUnknown() bool {
	return m.state == attr.ValueStateUnknown
}

// String implements basetypes.ObjectValuable.
func (m WebhookFilterInValue) String() string {
	panic("unimplemented")
}

// ToObjectValue implements basetypes.ObjectValuable.
func (m WebhookFilterInValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	panic("unimplemented")
}

// ToTerraformValue implements basetypes.ObjectValuable.
func (m WebhookFilterInValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	objectType := tftypes.Object{AttributeTypes: m.TerraformAttributeTypes(ctx)}

	return tftypes.NewValue(objectType, nil), nil
}

// Type implements basetypes.ObjectValuable.
func (m WebhookFilterInValue) Type(ctx context.Context) attr.Type {
	return WebhookFilterInType{
		ObjectType: m.ObjectType(ctx),
	}
}

func (m WebhookFilterInValue) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{AttributeTypes: m.TerraformAttributeTypes(ctx)}
}

func (m WebhookFilterInValue) TerraformAttributeTypes(ctx context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"doc":    tftypes.String,
		"values": tftypes.List{ElementType: tftypes.String},
	}
}

var _ basetypes.ObjectValuable = WebhookFilterInValue{}

func (m WebhookFilterInValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
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

func (m WebhookFilterInValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: m.AttributeTypes(ctx),
	}
}

func (m WebhookFilterInValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"doc": basetypes.StringType{},
		"values": basetypes.ListType{
			ElemType: basetypes.StringType{},
		},
	}
}
