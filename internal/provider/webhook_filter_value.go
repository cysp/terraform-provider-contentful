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

func NewWebhookFilterValueKnownFromAttributes(_ context.Context, attributes map[string]attr.Value) (WebhookFilterValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	notValue, notOk := attributes["not"].(WebhookFilterNotValue)
	if !notOk {
		diags.AddAttributeError(path.Root("not"), "invalid data", fmt.Sprintf("expected object of type WebhookFilterNotValue, got %T", attributes["not"]))
	}

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

	return WebhookFilterValue{
		Not:    notValue,
		Equals: equalsValue,
		In:     inValue,
		Regexp: regexpValue,
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

//nolint:ireturn
func (v WebhookFilterValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return WebhookFilterType{
		v.ObjectType(ctx),
	}
}

var _ basetypes.ObjectValuable = WebhookFilterValue{}

//nolint:ireturn
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
		"not":    WebhookFilterNotValue{}.CustomType(ctx),
		"equals": WebhookFilterEqualsValue{}.CustomType(ctx),
		"in":     WebhookFilterInValue{}.CustomType(ctx),
		"regexp": WebhookFilterRegexpValue{}.CustomType(ctx),
	}
}

func (v WebhookFilterValue) Equal(o attr.Value) bool {
	other, ok := o.(WebhookFilterValue)
	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state == attr.ValueStateKnown {
		return v.Not.Equal(other.Not) && v.Equals.Equal(other.Equals) && v.In.Equal(other.In) && v.Regexp.Equal(other.Regexp)
	}

	return true
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

	//nolint:gomnd,mnd
	val := make(map[string]tftypes.Value, 4)

	var notErr error
	val["not"], notErr = v.Not.ToTerraformValue(ctx)

	var equalsErr error
	val["equals"], equalsErr = v.Equals.ToTerraformValue(ctx)

	var inErr error
	val["in"], inErr = v.In.ToTerraformValue(ctx)

	var regexpErr error
	val["regexp"], regexpErr = v.Regexp.ToTerraformValue(ctx)

	validateErr := tftypes.ValidateValue(tft, val)

	err := errors.Join(notErr, equalsErr, inErr, regexpErr, validateErr)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(tft, val), nil
}

func (v WebhookFilterValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	attributeTypes := v.ObjectAttrTypes(ctx)

	switch {
	case v.IsNull():
		return types.ObjectNull(attributeTypes), nil
	case v.IsUnknown():
		return types.ObjectUnknown(attributeTypes), nil
	}

	attributes := map[string]attr.Value{
		"not":    v.Not,
		"equals": v.Equals,
		"in":     v.In,
		"regexp": v.Regexp,
	}

	return types.ObjectValue(attributeTypes, attributes)
}
