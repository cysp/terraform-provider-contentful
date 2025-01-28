package webhookfilter

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type WebhookFilterNotType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = WebhookFilterNotType{}

type WebhookFilterNotValue struct {
	Equals WebhookFilterEqualsValue `tfsdk:"equals"`
	In     WebhookFilterInValue     `tfsdk:"in"`
	Regexp WebhookFilterRegexpValue `tfsdk:"regexp"`
	state  attr.ValueState
}

func (v WebhookFilterNotValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return WebhookFilterNotType{
		ObjectType: v.ObjectType(ctx),
	}
}

var _ basetypes.ObjectValuable = WebhookFilterNotValue{}

// Equal implements basetypes.ObjectValuable.
func (m WebhookFilterNotValue) Equal(attr.Value) bool {
	panic("unimplemented")
}

// IsNull implements basetypes.ObjectValuable.
func (m WebhookFilterNotValue) IsNull() bool {
	return m.state == attr.ValueStateNull
}

// IsUnknown implements basetypes.ObjectValuable.
func (m WebhookFilterNotValue) IsUnknown() bool {
	return m.state == attr.ValueStateUnknown
}

// String implements basetypes.ObjectValuable.
func (m WebhookFilterNotValue) String() string {
	panic("unimplemented")
}

// ToObjectValue implements basetypes.ObjectValuable.
func (m WebhookFilterNotValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	panic("unimplemented")
}

// ToTerraformValue implements basetypes.ObjectValuable.
func (v WebhookFilterNotValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	objectType := tftypes.Object{AttributeTypes: v.TerraformAttributeTypes(ctx)}

	var val tftypes.Value
	var err error

	vals := make(map[string]tftypes.Value, 4)

	val, err = v.Equals.ToTerraformValue(ctx)

	if err != nil {
		return tftypes.NewValue(objectType, tftypes.UnknownValue), err
	}

	vals["equals"] = val

	val, err = v.In.ToTerraformValue(ctx)

	if err != nil {
		return tftypes.NewValue(objectType, tftypes.UnknownValue), err
	}

	vals["in"] = val

	val, err = v.Regexp.ToTerraformValue(ctx)

	if err != nil {
		return tftypes.NewValue(objectType, tftypes.UnknownValue), err
	}

	vals["regexp"] = val

	if err := tftypes.ValidateValue(objectType, vals); err != nil {
		return tftypes.NewValue(objectType, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(objectType, vals), nil
}

// Type implements basetypes.ObjectValuable.
func (m WebhookFilterNotValue) Type(context.Context) attr.Type {
	panic("unimplemented")
}

func (m WebhookFilterNotValue) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{AttributeTypes: m.TerraformAttributeTypes(ctx)}
}

func (m WebhookFilterNotValue) TerraformAttributeTypes(ctx context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"equals": WebhookFilterEqualsValue{}.TerraformType(ctx),
		"in":     WebhookFilterInValue{}.TerraformType(ctx),
		"regexp": WebhookFilterRegexpValue{}.TerraformType(ctx),
	}
}

var _ basetypes.ObjectValuable = WebhookFilterNotValue{}

func (m WebhookFilterNotValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"equals": schema.SingleNestedAttribute{
			Attributes: WebhookFilterEqualsValue{}.SchemaAttributes(ctx),
			CustomType: WebhookFilterEqualsValue{}.ObjectType(ctx),
			Optional:   true,
		},
		"in": schema.SingleNestedAttribute{
			Attributes: WebhookFilterInValue{}.SchemaAttributes(ctx),
			CustomType: WebhookFilterInValue{}.ObjectType(ctx),
			Optional:   true,
		},
		"regexp": schema.SingleNestedAttribute{
			Attributes: WebhookFilterRegexpValue{}.SchemaAttributes(ctx),
			CustomType: WebhookFilterRegexpValue{}.ObjectType(ctx),
			Optional:   true,
		},
	}
}

func (m WebhookFilterNotValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: m.AttributeTypes(ctx),
	}
}

func (m WebhookFilterNotValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"equals": WebhookFilterEqualsValue{}.ObjectType(ctx),
		"in":     WebhookFilterInValue{}.ObjectType(ctx),
		"regexp": WebhookFilterRegexpValue{}.ObjectType(ctx),
	}
}
