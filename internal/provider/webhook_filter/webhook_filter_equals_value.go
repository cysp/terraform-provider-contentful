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

//nolint:revive
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

func NewWebhookFilterEqualsValueKnownFromAttributes(_ context.Context, attributes map[string]attr.Value) (WebhookFilterEqualsValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	doc, docOk := attributes["doc"].(basetypes.StringValue)
	if !docOk {
		diags.AddError("Invalid data", fmt.Sprintf("expected doc to be of type String, got %T", attributes["doc"]))
	}

	value, valueOk := attributes["value"].(basetypes.StringValue)
	if !valueOk {
		diags.AddError("Invalid data", fmt.Sprintf("expected value to be of type String, got %T", attributes["doc"]))
	}

	return WebhookFilterEqualsValue{
		Doc:   doc,
		Value: value,
		state: attr.ValueStateKnown,
	}, diags
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
	return WebhookFilterEqualsType{
		v.ObjectType(ctx),
	}
}

var _ basetypes.ObjectValuable = WebhookFilterEqualsValue{}

//nolint:ireturn
func (v WebhookFilterEqualsValue) Type(ctx context.Context) attr.Type {
	return WebhookFilterEqualsType{
		ObjectType: v.ObjectType(ctx),
	}
}

func (v WebhookFilterEqualsValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: v.ObjectAttrTypes(ctx),
	}
}

func (v WebhookFilterEqualsValue) ObjectAttrTypes(_ context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"doc":   types.StringType,
		"value": types.StringType,
	}
}

func (v WebhookFilterEqualsValue) Equal(o attr.Value) bool {
	other, ok := o.(WebhookFilterEqualsValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Doc.Equal(other.Doc) {
		return false
	}

	if !v.Value.Equal(other.Value) {
		return false
	}

	return true
}

func (v WebhookFilterEqualsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v WebhookFilterEqualsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v WebhookFilterEqualsValue) String() string {
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

	var err error

	val := make(map[string]tftypes.Value, 2)

	val["doc"], err = v.Doc.ToTerraformValue(ctx)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	val["value"], err = v.Value.ToTerraformValue(ctx)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	if err := tftypes.ValidateValue(tft, val); err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(tft, val), nil
}

func (v WebhookFilterEqualsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributeTypes := v.ObjectAttrTypes(ctx)

	if v.IsNull() {
		return types.ObjectNull(attributeTypes), diags
	}

	if v.IsUnknown() {
		return types.ObjectUnknown(attributeTypes), diags
	}

	attributes := map[string]attr.Value{
		"doc":   v.Doc,
		"value": v.Value,
	}

	return types.ObjectValue(attributeTypes, attributes)
}
