package webhookfilter

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type WebhookFilterEqualsValue struct {
	Doc   basetypes.StringValue `tfsdk:"doc"`
	Value basetypes.StringValue `tfsdk:"value"`
	state attr.ValueState
}

func NewWebhookFilterEqualsValueKnown() WebhookFilterEqualsValue {
	return WebhookFilterEqualsValue{
		Doc:   basetypes.NewStringNull(),
		Value: basetypes.NewStringNull(),
		state: attr.ValueStateKnown,
	}
}

func NewWebhookFilterEqualsValueKnownFromAttributes(attributes map[string]attr.Value) WebhookFilterEqualsValue {

	return WebhookFilterEqualsValue{
		Doc:   attributes["doc"].(basetypes.StringValue),
		Value: attributes["value"].(basetypes.StringValue),
		state: attr.ValueStateKnown,
	}
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

func (m WebhookFilterEqualsValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"doc": schema.StringAttribute{
			Required: true,
		},
		"value": schema.StringAttribute{
			Required: true,
		},
	}
}

func (m WebhookFilterEqualsValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return WebhookFilterEqualsType{
		m.ObjectType(ctx),
	}
}

var _ basetypes.ObjectValuable = WebhookFilterEqualsValue{}

func (m WebhookFilterEqualsValue) Type(ctx context.Context) attr.Type {
	return WebhookFilterEqualsType{
		ObjectType: m.ObjectType(ctx),
	}
}

func (m WebhookFilterEqualsValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: m.ObjectAttrTypes(ctx),
	}
}

func (m WebhookFilterEqualsValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"doc":   types.StringType,
		"value": types.StringType,
	}
}

func (m WebhookFilterEqualsValue) Equal(other attr.Value) bool {
	return false
}

func (m WebhookFilterEqualsValue) IsNull() bool {
	return m.state == attr.ValueStateNull
}

func (m WebhookFilterEqualsValue) IsUnknown() bool {
	return m.state == attr.ValueStateUnknown
}

func (m WebhookFilterEqualsValue) String() string {
	return ""
}

func (v WebhookFilterEqualsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	tft := WebhookFilterEqualsType{}.TerraformType(ctx)

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

	var val tftypes.Value
	var err error

	vals := make(map[string]tftypes.Value, 4)

	val, err = v.Doc.ToTerraformValue(ctx)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	vals["doc"] = val

	val, err = v.Value.ToTerraformValue(ctx)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	vals["value"] = val

	if err := tftypes.ValidateValue(tft, vals); err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(tft, vals), nil
}

func (m WebhookFilterEqualsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributeTypes := m.ObjectAttrTypes(ctx)

	if m.IsNull() {
		return types.ObjectNull(attributeTypes), diags
	}

	if m.IsUnknown() {
		return types.ObjectUnknown(attributeTypes), diags
	}

	objVal, diags := types.ObjectValue(
		attributeTypes,
		map[string]attr.Value{
			"doc":   m.Doc,
			"value": m.Value,
		})

	return objVal, diags
}
