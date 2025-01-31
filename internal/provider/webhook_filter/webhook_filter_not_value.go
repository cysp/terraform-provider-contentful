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

	equals, equalsDiags := WebhookFilterEqualsType{}.ValueFromObject(ctx, attributes["equals"].(basetypes.ObjectValue))
	diags.Append(equalsDiags...)

	in, inDiags := WebhookFilterInType{}.ValueFromObject(ctx, attributes["in"].(basetypes.ObjectValue))
	diags.Append(inDiags...)

	regexp, regexpDiags := WebhookFilterRegexpType{}.ValueFromObject(ctx, attributes["regexp"].(basetypes.ObjectValue))
	diags.Append(regexpDiags...)

	return WebhookFilterNotValue{
		Equals: equals.(WebhookFilterEqualsValue),
		In:     in.(WebhookFilterInValue),
		Regexp: regexp.(WebhookFilterRegexpValue),
		state:  attr.ValueStateKnown,
	}, diags
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

//nolint:ireturn
func (v WebhookFilterNotValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return WebhookFilterNotType{
		v.ObjectType(ctx),
	}
}

var _ basetypes.ObjectValuable = WebhookFilterNotValue{}

//nolint:ireturn
func (v WebhookFilterNotValue) Type(ctx context.Context) attr.Type {
	return WebhookFilterNotType{
		ObjectType: v.ObjectType(ctx),
	}
}

func (v WebhookFilterNotValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: v.ObjectAttrTypes(ctx),
	}
}

func (v WebhookFilterNotValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"equals": WebhookFilterEqualsValue{}.ObjectType(ctx),
		"in":     WebhookFilterInValue{}.ObjectType(ctx),
		"regexp": WebhookFilterRegexpValue{}.ObjectType(ctx),
	}
}

func (v WebhookFilterNotValue) Equal(o attr.Value) bool {
	other, ok := o.(WebhookFilterNotValue)
	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Equals.Equal(other.Equals) {
		return false
	}

	if !v.In.Equal(other.In) {
		return false
	}

	if !v.Regexp.Equal(other.Regexp) {
		return false
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
	tft := WebhookFilterNotType{}.TerraformType(ctx)

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

	val := make(map[string]tftypes.Value, 4)

	val["equals"], err = v.Equals.ToTerraformValue(ctx)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	val["in"], err = v.In.ToTerraformValue(ctx)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	val["regexp"], err = v.Regexp.ToTerraformValue(ctx)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	if err := tftypes.ValidateValue(tft, val); err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(tft, val), nil
}

func (v WebhookFilterNotValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributeTypes := v.ObjectAttrTypes(ctx)

	if v.IsNull() {
		return types.ObjectNull(attributeTypes), diags
	}

	if v.IsUnknown() {
		return types.ObjectUnknown(attributeTypes), diags
	}

	attributes := map[string]attr.Value{
		"equals": v.Equals,
		"in":     v.In,
		"regexp": v.Regexp,
	}

	return types.ObjectValue(attributeTypes, attributes)
}
