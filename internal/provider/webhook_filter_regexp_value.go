package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type WebhookFilterRegexpValue struct {
	Doc     basetypes.StringValue `tfsdk:"doc"`
	Pattern basetypes.StringValue `tfsdk:"pattern"`
	state   attr.ValueState
}

func NewWebhookFilterRegexpValueKnown() WebhookFilterRegexpValue {
	return WebhookFilterRegexpValue{
		Doc:     basetypes.NewStringNull(),
		Pattern: basetypes.NewStringNull(),
		state:   attr.ValueStateKnown,
	}
}

func NewWebhookFilterRegexpValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (WebhookFilterRegexpValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := WebhookFilterRegexpValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := setTFSDKAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewWebhookFilterRegexpValueNull() WebhookFilterRegexpValue {
	return WebhookFilterRegexpValue{
		state: attr.ValueStateNull,
	}
}

func NewWebhookFilterRegexpValueUnknown() WebhookFilterRegexpValue {
	return WebhookFilterRegexpValue{
		state: attr.ValueStateUnknown,
	}
}

func (v WebhookFilterRegexpValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"doc": schema.StringAttribute{
			Required: true,
		},
		"pattern": schema.StringAttribute{
			Required: true,
		},
	}
}

//nolint:ireturn
func (v WebhookFilterRegexpValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return WebhookFilterRegexpType{
		v.ObjectType(ctx),
	}
}

var _ basetypes.ObjectValuable = WebhookFilterRegexpValue{}

//nolint:ireturn
func (v WebhookFilterRegexpValue) Type(ctx context.Context) attr.Type {
	return WebhookFilterRegexpType{
		ObjectType: v.ObjectType(ctx),
	}
}

func (v WebhookFilterRegexpValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: v.ObjectAttrTypes(ctx),
	}
}

func (v WebhookFilterRegexpValue) ObjectAttrTypes(_ context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"doc":     types.StringType,
		"pattern": types.StringType,
	}
}

func (v WebhookFilterRegexpValue) Equal(o attr.Value) bool {
	other, ok := o.(WebhookFilterRegexpValue)
	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return v.Doc.Equal(other.Doc) && v.Pattern.Equal(other.Pattern)
	}

	return true
}

func (v WebhookFilterRegexpValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v WebhookFilterRegexpValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v WebhookFilterRegexpValue) String() string {
	panic("unimplemented")
}

//nolint:dupl
func (v WebhookFilterRegexpValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	tft := WebhookFilterRegexpType{}.TerraformType(ctx)

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
	val := make(map[string]tftypes.Value, 2)

	var docErr error
	val["doc"], docErr = v.Doc.ToTerraformValue(ctx)

	var patternErr error
	val["pattern"], patternErr = v.Pattern.ToTerraformValue(ctx)

	validateErr := tftypes.ValidateValue(tft, val)

	err := errors.Join(docErr, patternErr, validateErr)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(tft, val), nil
}

func (v WebhookFilterRegexpValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	attributeTypes := v.ObjectAttrTypes(ctx)

	switch {
	case v.IsNull():
		return types.ObjectNull(attributeTypes), nil
	case v.IsUnknown():
		return types.ObjectUnknown(attributeTypes), nil
	}

	attributes := map[string]attr.Value{
		"doc":     v.Doc,
		"pattern": v.Pattern,
	}

	return types.ObjectValue(attributeTypes, attributes)
}
