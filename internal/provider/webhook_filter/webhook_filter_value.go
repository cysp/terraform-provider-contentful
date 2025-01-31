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

	not, notDiags := WebhookFilterNotType{}.ValueFromObject(ctx, attributes["not"].(basetypes.ObjectValue))
	diags.Append(notDiags...)

	equals, equalsDiags := WebhookFilterEqualsType{}.ValueFromObject(ctx, attributes["equals"].(basetypes.ObjectValue))
	diags.Append(equalsDiags...)

	in, inDiags := WebhookFilterInType{}.ValueFromObject(ctx, attributes["in"].(basetypes.ObjectValue))
	diags.Append(inDiags...)

	regexp, regexpDiags := WebhookFilterRegexpType{}.ValueFromObject(ctx, attributes["regexp"].(basetypes.ObjectValue))
	diags.Append(regexpDiags...)

	return WebhookFilterValue{
		Not:    not.(WebhookFilterNotValue),
		Equals: equals.(WebhookFilterEqualsValue),
		In:     in.(WebhookFilterInValue),
		Regexp: regexp.(WebhookFilterRegexpValue),
		state:  attr.ValueStateKnown,
	}, diags
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

func (v WebhookFilterValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"not": schema.SingleNestedAttribute{
			Attributes: WebhookFilterNotValue{}.SchemaAttributes(ctx),
			CustomType: WebhookFilterNotValue{}.CustomType(ctx),
			Optional:   true,
		},
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

func (v WebhookFilterValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return WebhookFilterType{
		v.ObjectType(ctx),
	}
}

var _ basetypes.ObjectValuable = WebhookFilterValue{}

func (v WebhookFilterValue) Type(ctx context.Context) attr.Type {
	return WebhookFilterType{
		ObjectType: v.ObjectType(ctx),
	}
}

func (v WebhookFilterValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: v.ObjectAttrTypes(ctx),
	}
}

func (v WebhookFilterValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"not":    WebhookFilterNotValue{}.ObjectType(ctx),
		"equals": WebhookFilterEqualsValue{}.ObjectType(ctx),
		"in":     WebhookFilterInValue{}.ObjectType(ctx),
		"regexp": WebhookFilterRegexpValue{}.ObjectType(ctx),
	}
}

func (v WebhookFilterValue) Equal(other attr.Value) bool {
	return false
}

func (v WebhookFilterValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v WebhookFilterValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v WebhookFilterValue) String() string {
	return ""
}

func (v WebhookFilterValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	tft := WebhookFilterType{}.TerraformType(ctx)

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

	val, err = v.Not.ToTerraformValue(ctx)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	vals["not"] = val

	val, err = v.Equals.ToTerraformValue(ctx)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	vals["equals"] = val

	val, err = v.In.ToTerraformValue(ctx)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	vals["in"] = val

	val, err = v.Regexp.ToTerraformValue(ctx)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	vals["regexp"] = val

	if err := tftypes.ValidateValue(tft, vals); err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(tft, vals), nil
}

func (v WebhookFilterValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributeTypes := v.ObjectAttrTypes(ctx)

	if v.IsNull() {
		return types.ObjectNull(attributeTypes), diags
	}

	if v.IsUnknown() {
		return types.ObjectUnknown(attributeTypes), diags
	}

	objVal, diags := types.ObjectValue(
		attributeTypes,
		map[string]attr.Value{
			"not":    v.Not,
			"equals": v.Equals,
			"in":     v.In,
			"regexp": v.Regexp,
		})

	return objVal, diags
}
