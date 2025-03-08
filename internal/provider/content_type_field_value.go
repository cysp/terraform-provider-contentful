package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

func NewFieldsValueNull() FieldsValue {
	return FieldsValue{
		state: attr.ValueStateNull,
	}
}

func NewFieldsValueUnknown() FieldsValue {
	return FieldsValue{
		state: attr.ValueStateUnknown,
	}
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

func (v FieldsValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return FieldsType{
		v.ObjectType(ctx),
	}
}

func (v FieldsValue) Type(ctx context.Context) attr.Type {
	return FieldsType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
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
		"disabled":      basetypes.BoolType{},
		"omitted":       basetypes.BoolType{},
		"required":      basetypes.BoolType{},
		"default_value": basetypes.StringType{},
		"items":         basetypes.ObjectType{AttrTypes: ItemsValue{}.AttributeTypes(ctx)},
		"link_type":     basetypes.StringType{},
		"localized":     basetypes.BoolType{},
		"validations":   basetypes.ListType{ElemType: types.StringType},
	}
}

func (v FieldsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 11)

	var val tftypes.Value
	var err error

	attrTypes["default_value"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["disabled"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["id"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["items"] = basetypes.ObjectType{
		AttrTypes: ItemsValue{}.AttributeTypes(ctx),
	}.TerraformType(ctx)
	attrTypes["link_type"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["localized"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["name"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["omitted"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["required"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["type"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["validations"] = basetypes.ListType{
		ElemType: types.StringType,
	}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 11)

		val, err = v.DefaultValue.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["default_value"] = val

		val, err = v.Disabled.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["disabled"] = val

		val, err = v.Id.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["id"] = val

		val, err = v.Items.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["items"] = val

		val, err = v.LinkType.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["link_type"] = val

		val, err = v.Localized.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["localized"] = val

		val, err = v.Name.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["name"] = val

		val, err = v.Omitted.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["omitted"] = val

		val, err = v.Required.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["required"] = val

		val, err = v.FieldsType.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["type"] = val

		val, err = v.Validations.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["validations"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
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

func (v FieldsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	var items basetypes.ObjectValue

	if v.Items.IsNull() {
		items = types.ObjectNull(
			ItemsValue{}.AttributeTypes(ctx),
		)
	}

	if v.Items.IsUnknown() {
		items = types.ObjectUnknown(
			ItemsValue{}.AttributeTypes(ctx),
		)
	}

	if !v.Items.IsNull() && !v.Items.IsUnknown() {
		items = types.ObjectValueMust(
			ItemsValue{}.AttributeTypes(ctx),
			v.Items.Attributes(),
		)
	}

	var validationsVal basetypes.ListValue
	switch {
	case v.Validations.IsUnknown():
		validationsVal = types.ListUnknown(types.StringType)
	case v.Validations.IsNull():
		validationsVal = types.ListNull(types.StringType)
	default:
		var d diag.Diagnostics
		validationsVal, d = types.ListValue(types.StringType, v.Validations.Elements())
		diags.Append(d...)
	}

	if diags.HasError() {
		return types.ObjectUnknown(map[string]attr.Type{
			"default_value": basetypes.StringType{},
			"disabled":      basetypes.BoolType{},
			"id":            basetypes.StringType{},
			"items": basetypes.ObjectType{
				AttrTypes: ItemsValue{}.AttributeTypes(ctx),
			},
			"link_type": basetypes.StringType{},
			"localized": basetypes.BoolType{},
			"name":      basetypes.StringType{},
			"omitted":   basetypes.BoolType{},
			"required":  basetypes.BoolType{},
			"type":      basetypes.StringType{},
			"validations": basetypes.ListType{
				ElemType: types.StringType,
			},
		}), diags
	}

	attributeTypes := map[string]attr.Type{
		"default_value": basetypes.StringType{},
		"disabled":      basetypes.BoolType{},
		"id":            basetypes.StringType{},
		"items": basetypes.ObjectType{
			AttrTypes: ItemsValue{}.AttributeTypes(ctx),
		},
		"link_type": basetypes.StringType{},
		"localized": basetypes.BoolType{},
		"name":      basetypes.StringType{},
		"omitted":   basetypes.BoolType{},
		"required":  basetypes.BoolType{},
		"type":      basetypes.StringType{},
		"validations": basetypes.ListType{
			ElemType: types.StringType,
		},
	}

	if v.IsNull() {
		return types.ObjectNull(attributeTypes), diags
	}

	if v.IsUnknown() {
		return types.ObjectUnknown(attributeTypes), diags
	}

	objVal, diags := types.ObjectValue(
		attributeTypes,
		map[string]attr.Value{
			"default_value": v.DefaultValue,
			"disabled":      v.Disabled,
			"id":            v.Id,
			"items":         items,
			"link_type":     v.LinkType,
			"localized":     v.Localized,
			"name":          v.Name,
			"omitted":       v.Omitted,
			"required":      v.Required,
			"type":          v.FieldsType,
			"validations":   validationsVal,
		})

	return objVal, diags
}

func (v FieldsValue) Equal(o attr.Value) bool {
	other, ok := o.(FieldsValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.DefaultValue.Equal(other.DefaultValue) {
		return false
	}

	if !v.Disabled.Equal(other.Disabled) {
		return false
	}

	if !v.Id.Equal(other.Id) {
		return false
	}

	if !v.Items.Equal(other.Items) {
		return false
	}

	if !v.LinkType.Equal(other.LinkType) {
		return false
	}

	if !v.Localized.Equal(other.Localized) {
		return false
	}

	if !v.Name.Equal(other.Name) {
		return false
	}

	if !v.Omitted.Equal(other.Omitted) {
		return false
	}

	if !v.Required.Equal(other.Required) {
		return false
	}

	if !v.FieldsType.Equal(other.FieldsType) {
		return false
	}

	if !v.Validations.Equal(other.Validations) {
		return false
	}

	return true
}

func (v FieldsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"id":            basetypes.StringType{},
		"default_value": basetypes.StringType{},
		"disabled":      basetypes.BoolType{},
		"items": basetypes.ObjectType{
			AttrTypes: ItemsValue{}.AttributeTypes(ctx),
		},
		"link_type": basetypes.StringType{},
		"localized": basetypes.BoolType{},
		"name":      basetypes.StringType{},
		"omitted":   basetypes.BoolType{},
		"required":  basetypes.BoolType{},
		"type":      basetypes.StringType{},
		"validations": basetypes.ListType{
			ElemType: types.StringType,
		},
	}
}
