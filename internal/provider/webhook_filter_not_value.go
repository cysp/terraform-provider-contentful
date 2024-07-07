package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

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

func NewWebhookFilterNotValueKnownFromAttributes(_ context.Context, attributes map[string]attr.Value) (WebhookFilterNotValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	equalsValue, equalsOk := attributes["equals"].(WebhookFilterEqualsValue)
	if !equalsOk {
		diags.AddAttributeError(path.Root("equals"), "invalid data", fmt.Sprintf("expected object of type WebhookFilterEqualsValue, got %T", attributes["equals"]))
	}

	inValue, inOk := attributes["in"].(WebhookFilterInValue)
	if !inOk {
		diags.AddAttributeError(path.Root("in"), "invalid data", fmt.Sprintf("expected object of type WebhookFilterInValue, got %T", attributes["in"]))
	}

	regexpValue, regexpOk := attributes["regexp"].(WebhookFilterRegexpValue)
	if !regexpOk {
		diags.AddAttributeError(path.Root("regexp"), "invalid data", fmt.Sprintf("expected object of type WebhookFilterRegexpValue, got %T", attributes["regexp"]))
	}

	return WebhookFilterNotValue{
		Equals: equalsValue,
		In:     inValue,
		Regexp: regexpValue,
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
		"equals": WebhookFilterEqualsValue{}.CustomType(ctx),
		"in":     WebhookFilterInValue{}.CustomType(ctx),
		"regexp": WebhookFilterRegexpValue{}.CustomType(ctx),
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

	if v.state == attr.ValueStateKnown {
		return v.Equals.Equal(other.Equals) && v.In.Equal(other.In) && v.Regexp.Equal(other.Regexp)
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

	//nolint:gomnd,mnd
	val := make(map[string]tftypes.Value, 3)

	var equalsErr error
	val["equals"], equalsErr = v.Equals.ToTerraformValue(ctx)

	var inErr error
	val["in"], inErr = v.In.ToTerraformValue(ctx)

	var regexpErr error
	val["regexp"], regexpErr = v.Regexp.ToTerraformValue(ctx)

	validateErr := tftypes.ValidateValue(tft, val)

	err := errors.Join(equalsErr, inErr, regexpErr, validateErr)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(tft, val), nil
}

func (v WebhookFilterNotValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	attributeTypes := v.ObjectAttrTypes(ctx)

	switch {
	case v.IsNull():
		return types.ObjectNull(attributeTypes), nil
	case v.IsUnknown():
		return types.ObjectUnknown(attributeTypes), nil
	}

	attributes := map[string]attr.Value{
		"equals": v.Equals,
		"in":     v.In,
		"regexp": v.Regexp,
	}

	return types.ObjectValue(attributeTypes, attributes)
}
