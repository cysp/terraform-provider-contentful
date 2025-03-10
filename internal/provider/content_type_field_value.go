package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

//nolint:recvcheck
type ContentTypeFieldValue struct {
	ID           basetypes.StringValue `tfsdk:"id"`
	Name         basetypes.StringValue `tfsdk:"name"`
	FieldType    basetypes.StringValue `tfsdk:"type"`
	LinkType     basetypes.StringValue `tfsdk:"link_type"`
	Disabled     basetypes.BoolValue   `tfsdk:"disabled"`
	Omitted      basetypes.BoolValue   `tfsdk:"omitted"`
	Required     basetypes.BoolValue   `tfsdk:"required"`
	DefaultValue jsontypes.Normalized  `tfsdk:"default_value"`
	Items        basetypes.ObjectValue `tfsdk:"items"`
	Localized    basetypes.BoolValue   `tfsdk:"localized"`
	Validations  basetypes.ListValue   `tfsdk:"validations"`
	state        attr.ValueState
}

var _ basetypes.ObjectValuable = ContentTypeFieldValue{}

func NewContentTypeFieldValueUnknown() ContentTypeFieldValue {
	return ContentTypeFieldValue{
		state: attr.ValueStateUnknown,
	}
}

func NewContentTypeFieldValueNull() ContentTypeFieldValue {
	return ContentTypeFieldValue{
		state: attr.ValueStateNull,
	}
}

func NewContentTypeFieldValueKnownFromAttributesMust(ctx context.Context, attributes map[string]attr.Value) ContentTypeFieldValue {
	value, diags := NewContentTypeFieldValueKnownFromAttributes(ctx, attributes)
	if diags.HasError() {
		panic(diags)
	}

	return value
}

func NewContentTypeFieldValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (ContentTypeFieldValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := ContentTypeFieldValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := setTFSDKAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func (v ContentTypeFieldValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
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
		"items": schema.SingleNestedAttribute{
			Attributes: ContentTypeFieldItemsValue{}.SchemaAttributes(ctx),
			CustomType: ContentTypeFieldItemsValue{}.CustomType(ctx),
			Optional:   true,
		},
		"default_value": schema.StringAttribute{
			CustomType: jsontypes.NormalizedType{},
			Optional:   true,
			Computed:   true,
		},
		"localized": schema.BoolAttribute{
			Required: true,
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
		"validations": schema.ListAttribute{
			ElementType: jsontypes.NormalizedType{},
			Optional:    true,
			Computed:    true,
			Default:     listdefault.StaticValue(NewEmptyListMust(jsontypes.NormalizedType{})),
		},
	}
}

//nolint:ireturn
func (v ContentTypeFieldValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return ContentTypeFieldType{
		v.ObjectType(ctx),
	}
}

//nolint:ireturn
func (v ContentTypeFieldValue) Type(ctx context.Context) attr.Type {
	return ContentTypeFieldType{
		ObjectType: v.ObjectType(ctx),
	}
}

func (v ContentTypeFieldValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: v.ObjectAttrTypes(ctx),
	}
}

func (v ContentTypeFieldValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"id":            basetypes.StringType{},
		"name":          basetypes.StringType{},
		"type":          basetypes.StringType{},
		"link_type":     basetypes.StringType{},
		"items":         ContentTypeFieldItemsValue{}.ObjectType(ctx),
		"default_value": jsontypes.NormalizedType{},
		"localized":     basetypes.BoolType{},
		"disabled":      basetypes.BoolType{},
		"omitted":       basetypes.BoolType{},
		"required":      basetypes.BoolType{},
		"validations":   basetypes.ListType{ElemType: jsontypes.NormalizedType{}},
	}
}

func (v ContentTypeFieldValue) Equal(o attr.Value) bool {
	other, ok := o.(ContentTypeFieldValue)
	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state == attr.ValueStateKnown {
		return v.ID.Equal(other.ID) &&
			v.Name.Equal(other.Name) &&
			v.FieldType.Equal(other.FieldType) &&
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

func (v ContentTypeFieldValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ContentTypeFieldValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ContentTypeFieldValue) String() string {
	return "ContentTypeFieldValue"
}

func (v ContentTypeFieldValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	tft := ContentTypeFieldType{}.TerraformType(ctx)

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
	val["id"], idErr = v.ID.ToTerraformValue(ctx)

	var nameErr error
	val["name"], nameErr = v.Name.ToTerraformValue(ctx)

	var typeErr error
	val["type"], typeErr = v.FieldType.ToTerraformValue(ctx)

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

func (v ContentTypeFieldValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	attributeTypes := v.ObjectAttrTypes(ctx)

	switch {
	case v.IsNull():
		return types.ObjectNull(attributeTypes), nil
	case v.IsUnknown():
		return types.ObjectUnknown(attributeTypes), nil
	}

	attributes := map[string]attr.Value{
		"id":            v.ID,
		"name":          v.Name,
		"type":          v.FieldType,
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
