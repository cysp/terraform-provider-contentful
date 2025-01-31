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

func NewWebhookFilterInValueKnownFromAttributes(_ context.Context, attributes map[string]attr.Value) (WebhookFilterInValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	doc, docOk := attributes["doc"].(basetypes.StringValue)
	if !docOk {
		diags.AddError("Invalid data", fmt.Sprintf("expected doc to be of type String, got %T", attributes["doc"]))
	}

	values, valueOk := attributes["values"].(basetypes.ListValue)
	if !valueOk {
		diags.AddError("Invalid data", fmt.Sprintf("expected value to be of type List[String], got %T", attributes["doc"]))
	}

	return WebhookFilterInValue{
		Doc:    doc,
		Values: values,
		state:  attr.ValueStateKnown,
	}, diags
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
	return WebhookFilterInType{
		v.ObjectType(ctx),
	}
}

var _ basetypes.ObjectValuable = WebhookFilterInValue{}

//nolint:ireturn
func (v WebhookFilterInValue) Type(ctx context.Context) attr.Type {
	return WebhookFilterInType{
		ObjectType: v.ObjectType(ctx),
	}
}

func (v WebhookFilterInValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: v.ObjectAttrTypes(ctx),
	}
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

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Doc.Equal(other.Doc) {
		return false
	}

	if !v.Values.Equal(other.Values) {
		return false
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
	tft := WebhookFilterInType{}.TerraformType(ctx)

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

	val["values"], err = v.Values.ToTerraformValue(ctx)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	if err := tftypes.ValidateValue(tft, val); err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(tft, val), nil
}

func (v WebhookFilterInValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributeTypes := v.ObjectAttrTypes(ctx)

	if v.IsNull() {
		return types.ObjectNull(attributeTypes), diags
	}

	if v.IsUnknown() {
		return types.ObjectUnknown(attributeTypes), diags
	}

	attributes := map[string]attr.Value{
		"doc":    v.Doc,
		"values": v.Values,
	}

	return types.ObjectValue(attributeTypes, attributes)
}
