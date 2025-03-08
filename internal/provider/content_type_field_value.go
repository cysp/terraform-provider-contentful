package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type FieldsValue struct {
	Id           basetypes.StringValue `tfsdk:"id"`
	Name         basetypes.StringValue `tfsdk:"name"`
	FieldsType   basetypes.StringValue `tfsdk:"type"`
	LinkType     basetypes.StringValue `tfsdk:"link_type"`
	Disabled     basetypes.BoolValue   `tfsdk:"disabled"`
	Omitted      basetypes.BoolValue   `tfsdk:"omitted"`
	Required     basetypes.BoolValue   `tfsdk:"required"`
	DefaultValue basetypes.StringValue `tfsdk:"default_value"`
	Items        basetypes.ObjectValue `tfsdk:"items"`
	Localized    basetypes.BoolValue   `tfsdk:"localized"`
	Validations  basetypes.ListValue   `tfsdk:"validations"`
	state        attr.ValueState
}

var _ basetypes.ObjectValuable = FieldsValue{}

func NewFieldsValueUnknown() FieldsValue {
	return FieldsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewFieldsValueNull() FieldsValue {
	return FieldsValue{
		state: attr.ValueStateNull,
	}
}

func NewContentTypeFieldValueKnownFromAttributes(_ context.Context, attributes map[string]attr.Value) (WebhookFilterValue, diag.Diagnostics) {
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

func (v FieldsValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Required: true,
		},
		"name": schema.StringAttribute{
			Required: true,
		},
		"type": schema.StringAttribute{
			Required: true,
		},
		"link_type": schema.StringAttribute{
			Optional: true,
		},
		"disabled": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
		},
		"omitted": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
		},
		"required": schema.BoolAttribute{
			Required: true,
		},
		"default_value": schema.StringAttribute{
			CustomType: jsontypes.NormalizedType{},
			Optional:   true,
			Computed:   true,
			Default:    stringdefault.StaticString(""),
		},
		"items": schema.SingleNestedAttribute{
			Attributes: map[string]schema.Attribute{
				"type": schema.StringAttribute{
					Required: true,
				},
				"link_type": schema.StringAttribute{
					Optional: true,
				},
				"validations": schema.ListAttribute{
					ElementType: jsontypes.NormalizedType{},
					Optional:    true,
					Computed:    true,
					Default:     listdefault.StaticValue(NewEmptyListMust(jsontypes.NormalizedType{})),
				},
			},
			CustomType: ItemsType{
				ObjectType: types.ObjectType{
					AttrTypes: ItemsValue{}.AttributeTypes(ctx),
				},
			},
			Optional: true,
		},
		"localized": schema.BoolAttribute{
			Required: true,
		},
		"validations": schema.ListAttribute{
			ElementType: jsontypes.NormalizedType{},
			Optional:    true,
			Computed:    true,
			Default:     listdefault.StaticValue(NewEmptyListMust(jsontypes.NormalizedType{})),
		},
	}
}

//nolint:ireturn
func (v FieldsValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return FieldsType{
		v.ObjectType(ctx),
	}
}

//nolint:ireturn
func (v FieldsValue) Type(ctx context.Context) attr.Type {
	return FieldsType{
		ObjectType: v.ObjectType(ctx),
	}
}

func (v FieldsValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: v.ObjectAttrTypes(ctx),
	}
}

func (v FieldsValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"id":            basetypes.StringType{},
		"name":          basetypes.StringType{},
		"type":          basetypes.StringType{},
		"link_type":     basetypes.StringType{},
		"items":         basetypes.ObjectType{AttrTypes: ItemsValue{}.AttributeTypes(ctx)},
		"default_value": basetypes.StringType{},
		"localized":     basetypes.BoolType{},
		"disabled":      basetypes.BoolType{},
		"omitted":       basetypes.BoolType{},
		"required":      basetypes.BoolType{},
		"validations":   basetypes.ListType{ElemType: types.StringType},
	}
}

func (v FieldsValue) Equal(o attr.Value) bool {
	other, ok := o.(FieldsValue)
	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state == attr.ValueStateKnown {
		return v.Id.Equal(other.Id) &&
			v.Name.Equal(other.Name) &&
			v.FieldsType.Equal(other.FieldsType) &&
			v.LinkType.Equal(other.LinkType) &&
			v.Items.Equal(other.Items) &&
			v.DefaultValue.Equal(other.DefaultValue) &&
			v.Localized.Equal(other.Localized) &&
			v.Disabled.Equal(other.Disabled) &&
			v.Omitted.Equal(other.Omitted) &&
			v.Required.Equal(other.Required) &&
			v.Validations.Equal(other.Validations)
	}

	return true
}

func (v FieldsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v FieldsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v FieldsValue) String() string {
	return "FieldsValue"
}

func (v FieldsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	tft := FieldsType{}.TerraformType(ctx)

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
	val := make(map[string]tftypes.Value, 11)

	var idErr error
	val["id"], idErr = v.Id.ToTerraformValue(ctx)

	var nameErr error
	val["name"], nameErr = v.Name.ToTerraformValue(ctx)

	var typeErr error
	val["type"], typeErr = v.FieldsType.ToTerraformValue(ctx)

	var linkTypeErr error
	val["link_type"], linkTypeErr = v.LinkType.ToTerraformValue(ctx)

	var itemsErr error
	val["items"], itemsErr = v.Items.ToTerraformValue(ctx)

	var defaultValueErr error
	val["default_value"], defaultValueErr = v.DefaultValue.ToTerraformValue(ctx)

	var localizedErr error
	val["localized"], localizedErr = v.Localized.ToTerraformValue(ctx)

	var disabledErr error
	val["disabled"], disabledErr = v.Disabled.ToTerraformValue(ctx)

	var omittedErr error
	val["omitted"], omittedErr = v.Omitted.ToTerraformValue(ctx)

	var requiredErr error
	val["required"], requiredErr = v.Required.ToTerraformValue(ctx)

	var validationsErr error
	val["validations"], validationsErr = v.Validations.ToTerraformValue(ctx)

	validateErr := tftypes.ValidateValue(tft, val)

	err := errors.Join(idErr, nameErr, typeErr, linkTypeErr, itemsErr, defaultValueErr, localizedErr, disabledErr, omittedErr, requiredErr, validationsErr, validateErr)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(tft, val), nil
}

func (v FieldsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	attributeTypes := v.ObjectAttrTypes(ctx)

	switch {
	case v.IsNull():
		return types.ObjectNull(attributeTypes), nil
	case v.IsUnknown():
		return types.ObjectUnknown(attributeTypes), nil
	}

	attributes := map[string]attr.Value{
		"id":            v.Id,
		"name":          v.Name,
		"type":          v.FieldsType,
		"link_type":     v.LinkType,
		"items":         v.Items,
		"default_value": v.DefaultValue,
		"localized":     v.Localized,
		"disabled":      v.Disabled,
		"omitted":       v.Omitted,
		"required":      v.Required,
		"validations":   v.Validations,
	}

	return types.ObjectValue(attributeTypes, attributes)
}
