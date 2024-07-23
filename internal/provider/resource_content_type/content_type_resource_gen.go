// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_content_type

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func ContentTypeResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"content_type_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Required: true,
			},
			"display_field": schema.StringAttribute{
				Required: true,
			},
			"environment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"fields": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"default_value": schema.StringAttribute{
							CustomType: jsontypes.NormalizedType{},
							Optional:   true,
							Computed:   true,
						},
						"disabled": schema.BoolAttribute{
							Optional: true,
							Computed: true,
						},
						"id": schema.StringAttribute{
							Required: true,
						},
						"items": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"link_type": schema.StringAttribute{
									Optional: true,
								},
								"type": schema.StringAttribute{
									Required: true,
								},
								"validations": schema.ListAttribute{
									ElementType: jsontypes.NormalizedType{},
									Optional:    true,
									Computed:    true,
								},
							},
							CustomType: ItemsType{
								ObjectType: types.ObjectType{
									AttrTypes: ItemsValue{}.AttributeTypes(ctx),
								},
							},
							Optional: true,
						},
						"link_type": schema.StringAttribute{
							Optional: true,
						},
						"localized": schema.BoolAttribute{
							Required: true,
						},
						"name": schema.StringAttribute{
							Required: true,
						},
						"omitted": schema.BoolAttribute{
							Optional: true,
							Computed: true,
						},
						"required": schema.BoolAttribute{
							Required: true,
						},
						"type": schema.StringAttribute{
							Required: true,
						},
						"validations": schema.ListAttribute{
							ElementType: jsontypes.NormalizedType{},
							Optional:    true,
							Computed:    true,
						},
					},
					CustomType: FieldsType{
						ObjectType: types.ObjectType{
							AttrTypes: FieldsValue{}.AttributeTypes(ctx),
						},
					},
				},
				Required: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"space_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

type ContentTypeModel struct {
	ContentTypeId types.String `tfsdk:"content_type_id"`
	Description   types.String `tfsdk:"description"`
	DisplayField  types.String `tfsdk:"display_field"`
	EnvironmentId types.String `tfsdk:"environment_id"`
	Fields        types.List   `tfsdk:"fields"`
	Name          types.String `tfsdk:"name"`
	SpaceId       types.String `tfsdk:"space_id"`
}

var _ basetypes.ObjectTypable = FieldsType{}

type FieldsType struct {
	basetypes.ObjectType
}

func (t FieldsType) Equal(o attr.Type) bool {
	other, ok := o.(FieldsType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t FieldsType) String() string {
	return "FieldsType"
}

func (t FieldsType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	defaultValueAttribute, ok := attributes["default_value"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`default_value is missing from object`)

		return nil, diags
	}

	defaultValueVal, ok := defaultValueAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`default_value expected to be basetypes.StringValue, was: %T`, defaultValueAttribute))
	}

	disabledAttribute, ok := attributes["disabled"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`disabled is missing from object`)

		return nil, diags
	}

	disabledVal, ok := disabledAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`disabled expected to be basetypes.BoolValue, was: %T`, disabledAttribute))
	}

	idAttribute, ok := attributes["id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`id is missing from object`)

		return nil, diags
	}

	idVal, ok := idAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`id expected to be basetypes.StringValue, was: %T`, idAttribute))
	}

	itemsAttribute, ok := attributes["items"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`items is missing from object`)

		return nil, diags
	}

	itemsVal, ok := itemsAttribute.(basetypes.ObjectValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`items expected to be basetypes.ObjectValue, was: %T`, itemsAttribute))
	}

	linkTypeAttribute, ok := attributes["link_type"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`link_type is missing from object`)

		return nil, diags
	}

	linkTypeVal, ok := linkTypeAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`link_type expected to be basetypes.StringValue, was: %T`, linkTypeAttribute))
	}

	localizedAttribute, ok := attributes["localized"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`localized is missing from object`)

		return nil, diags
	}

	localizedVal, ok := localizedAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`localized expected to be basetypes.BoolValue, was: %T`, localizedAttribute))
	}

	nameAttribute, ok := attributes["name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`name is missing from object`)

		return nil, diags
	}

	nameVal, ok := nameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`name expected to be basetypes.StringValue, was: %T`, nameAttribute))
	}

	omittedAttribute, ok := attributes["omitted"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`omitted is missing from object`)

		return nil, diags
	}

	omittedVal, ok := omittedAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`omitted expected to be basetypes.BoolValue, was: %T`, omittedAttribute))
	}

	requiredAttribute, ok := attributes["required"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`required is missing from object`)

		return nil, diags
	}

	requiredVal, ok := requiredAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`required expected to be basetypes.BoolValue, was: %T`, requiredAttribute))
	}

	typeAttribute, ok := attributes["type"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`type is missing from object`)

		return nil, diags
	}

	typeVal, ok := typeAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`type expected to be basetypes.StringValue, was: %T`, typeAttribute))
	}

	validationsAttribute, ok := attributes["validations"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`validations is missing from object`)

		return nil, diags
	}

	validationsVal, ok := validationsAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`validations expected to be basetypes.ListValue, was: %T`, validationsAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return FieldsValue{
		DefaultValue: defaultValueVal,
		Disabled:     disabledVal,
		Id:           idVal,
		Items:        itemsVal,
		LinkType:     linkTypeVal,
		Localized:    localizedVal,
		Name:         nameVal,
		Omitted:      omittedVal,
		Required:     requiredVal,
		FieldsType:   typeVal,
		Validations:  validationsVal,
		state:        attr.ValueStateKnown,
	}, diags
}

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

func NewFieldsValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (FieldsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing FieldsValue Attribute Value",
				"While creating a FieldsValue value, a missing attribute value was detected. "+
					"A FieldsValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("FieldsValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid FieldsValue Attribute Type",
				"While creating a FieldsValue value, an invalid attribute value was detected. "+
					"A FieldsValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("FieldsValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("FieldsValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra FieldsValue Attribute Value",
				"While creating a FieldsValue value, an extra attribute value was detected. "+
					"A FieldsValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra FieldsValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewFieldsValueUnknown(), diags
	}

	defaultValueAttribute, ok := attributes["default_value"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`default_value is missing from object`)

		return NewFieldsValueUnknown(), diags
	}

	defaultValueVal, ok := defaultValueAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`default_value expected to be basetypes.StringValue, was: %T`, defaultValueAttribute))
	}

	disabledAttribute, ok := attributes["disabled"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`disabled is missing from object`)

		return NewFieldsValueUnknown(), diags
	}

	disabledVal, ok := disabledAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`disabled expected to be basetypes.BoolValue, was: %T`, disabledAttribute))
	}

	idAttribute, ok := attributes["id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`id is missing from object`)

		return NewFieldsValueUnknown(), diags
	}

	idVal, ok := idAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`id expected to be basetypes.StringValue, was: %T`, idAttribute))
	}

	itemsAttribute, ok := attributes["items"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`items is missing from object`)

		return NewFieldsValueUnknown(), diags
	}

	itemsVal, ok := itemsAttribute.(basetypes.ObjectValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`items expected to be basetypes.ObjectValue, was: %T`, itemsAttribute))
	}

	linkTypeAttribute, ok := attributes["link_type"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`link_type is missing from object`)

		return NewFieldsValueUnknown(), diags
	}

	linkTypeVal, ok := linkTypeAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`link_type expected to be basetypes.StringValue, was: %T`, linkTypeAttribute))
	}

	localizedAttribute, ok := attributes["localized"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`localized is missing from object`)

		return NewFieldsValueUnknown(), diags
	}

	localizedVal, ok := localizedAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`localized expected to be basetypes.BoolValue, was: %T`, localizedAttribute))
	}

	nameAttribute, ok := attributes["name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`name is missing from object`)

		return NewFieldsValueUnknown(), diags
	}

	nameVal, ok := nameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`name expected to be basetypes.StringValue, was: %T`, nameAttribute))
	}

	omittedAttribute, ok := attributes["omitted"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`omitted is missing from object`)

		return NewFieldsValueUnknown(), diags
	}

	omittedVal, ok := omittedAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`omitted expected to be basetypes.BoolValue, was: %T`, omittedAttribute))
	}

	requiredAttribute, ok := attributes["required"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`required is missing from object`)

		return NewFieldsValueUnknown(), diags
	}

	requiredVal, ok := requiredAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`required expected to be basetypes.BoolValue, was: %T`, requiredAttribute))
	}

	typeAttribute, ok := attributes["type"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`type is missing from object`)

		return NewFieldsValueUnknown(), diags
	}

	typeVal, ok := typeAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`type expected to be basetypes.StringValue, was: %T`, typeAttribute))
	}

	validationsAttribute, ok := attributes["validations"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`validations is missing from object`)

		return NewFieldsValueUnknown(), diags
	}

	validationsVal, ok := validationsAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`validations expected to be basetypes.ListValue, was: %T`, validationsAttribute))
	}

	if diags.HasError() {
		return NewFieldsValueUnknown(), diags
	}

	return FieldsValue{
		DefaultValue: defaultValueVal,
		Disabled:     disabledVal,
		Id:           idVal,
		Items:        itemsVal,
		LinkType:     linkTypeVal,
		Localized:    localizedVal,
		Name:         nameVal,
		Omitted:      omittedVal,
		Required:     requiredVal,
		FieldsType:   typeVal,
		Validations:  validationsVal,
		state:        attr.ValueStateKnown,
	}, diags
}

func NewFieldsValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) FieldsValue {
	object, diags := NewFieldsValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewFieldsValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t FieldsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewFieldsValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewFieldsValueUnknown(), nil
	}

	if in.IsNull() {
		return NewFieldsValueNull(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := in.As(&val)

	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return NewFieldsValueMust(FieldsValue{}.AttributeTypes(ctx), attributes), nil
}

func (t FieldsType) ValueType(ctx context.Context) attr.Value {
	return FieldsValue{}
}

var _ basetypes.ObjectValuable = FieldsValue{}

type FieldsValue struct {
	DefaultValue basetypes.StringValue `tfsdk:"default_value"`
	Disabled     basetypes.BoolValue   `tfsdk:"disabled"`
	Id           basetypes.StringValue `tfsdk:"id"`
	Items        basetypes.ObjectValue `tfsdk:"items"`
	LinkType     basetypes.StringValue `tfsdk:"link_type"`
	Localized    basetypes.BoolValue   `tfsdk:"localized"`
	Name         basetypes.StringValue `tfsdk:"name"`
	Omitted      basetypes.BoolValue   `tfsdk:"omitted"`
	Required     basetypes.BoolValue   `tfsdk:"required"`
	FieldsType   basetypes.StringValue `tfsdk:"type"`
	Validations  basetypes.ListValue   `tfsdk:"validations"`
	state        attr.ValueState
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

	validationsVal, d := types.ListValue(types.StringType, v.Validations.Elements())

	diags.Append(d...)

	if d.HasError() {
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

func (v FieldsValue) Type(ctx context.Context) attr.Type {
	return FieldsType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v FieldsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
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
}

var _ basetypes.ObjectTypable = ItemsType{}

type ItemsType struct {
	basetypes.ObjectType
}

func (t ItemsType) Equal(o attr.Type) bool {
	other, ok := o.(ItemsType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t ItemsType) String() string {
	return "ItemsType"
}

func (t ItemsType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	linkTypeAttribute, ok := attributes["link_type"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`link_type is missing from object`)

		return nil, diags
	}

	linkTypeVal, ok := linkTypeAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`link_type expected to be basetypes.StringValue, was: %T`, linkTypeAttribute))
	}

	typeAttribute, ok := attributes["type"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`type is missing from object`)

		return nil, diags
	}

	typeVal, ok := typeAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`type expected to be basetypes.StringValue, was: %T`, typeAttribute))
	}

	validationsAttribute, ok := attributes["validations"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`validations is missing from object`)

		return nil, diags
	}

	validationsVal, ok := validationsAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`validations expected to be basetypes.ListValue, was: %T`, validationsAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return ItemsValue{
		LinkType:    linkTypeVal,
		ItemsType:   typeVal,
		Validations: validationsVal,
		state:       attr.ValueStateKnown,
	}, diags
}

func NewItemsValueNull() ItemsValue {
	return ItemsValue{
		state: attr.ValueStateNull,
	}
}

func NewItemsValueUnknown() ItemsValue {
	return ItemsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewItemsValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (ItemsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing ItemsValue Attribute Value",
				"While creating a ItemsValue value, a missing attribute value was detected. "+
					"A ItemsValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("ItemsValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid ItemsValue Attribute Type",
				"While creating a ItemsValue value, an invalid attribute value was detected. "+
					"A ItemsValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("ItemsValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("ItemsValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra ItemsValue Attribute Value",
				"While creating a ItemsValue value, an extra attribute value was detected. "+
					"A ItemsValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra ItemsValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewItemsValueUnknown(), diags
	}

	linkTypeAttribute, ok := attributes["link_type"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`link_type is missing from object`)

		return NewItemsValueUnknown(), diags
	}

	linkTypeVal, ok := linkTypeAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`link_type expected to be basetypes.StringValue, was: %T`, linkTypeAttribute))
	}

	typeAttribute, ok := attributes["type"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`type is missing from object`)

		return NewItemsValueUnknown(), diags
	}

	typeVal, ok := typeAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`type expected to be basetypes.StringValue, was: %T`, typeAttribute))
	}

	validationsAttribute, ok := attributes["validations"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`validations is missing from object`)

		return NewItemsValueUnknown(), diags
	}

	validationsVal, ok := validationsAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`validations expected to be basetypes.ListValue, was: %T`, validationsAttribute))
	}

	if diags.HasError() {
		return NewItemsValueUnknown(), diags
	}

	return ItemsValue{
		LinkType:    linkTypeVal,
		ItemsType:   typeVal,
		Validations: validationsVal,
		state:       attr.ValueStateKnown,
	}, diags
}

func NewItemsValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) ItemsValue {
	object, diags := NewItemsValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewItemsValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t ItemsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewItemsValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewItemsValueUnknown(), nil
	}

	if in.IsNull() {
		return NewItemsValueNull(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := in.As(&val)

	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return NewItemsValueMust(ItemsValue{}.AttributeTypes(ctx), attributes), nil
}

func (t ItemsType) ValueType(ctx context.Context) attr.Value {
	return ItemsValue{}
}

var _ basetypes.ObjectValuable = ItemsValue{}

type ItemsValue struct {
	LinkType    basetypes.StringValue `tfsdk:"link_type"`
	ItemsType   basetypes.StringValue `tfsdk:"type"`
	Validations basetypes.ListValue   `tfsdk:"validations"`
	state       attr.ValueState
}

func (v ItemsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 3)

	var val tftypes.Value
	var err error

	attrTypes["link_type"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["type"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["validations"] = basetypes.ListType{
		ElemType: types.StringType,
	}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 3)

		val, err = v.LinkType.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["link_type"] = val

		val, err = v.ItemsType.ToTerraformValue(ctx)

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

func (v ItemsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ItemsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ItemsValue) String() string {
	return "ItemsValue"
}

func (v ItemsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	validationsVal, d := types.ListValue(types.StringType, v.Validations.Elements())

	diags.Append(d...)

	if d.HasError() {
		return types.ObjectUnknown(map[string]attr.Type{
			"link_type": basetypes.StringType{},
			"type":      basetypes.StringType{},
			"validations": basetypes.ListType{
				ElemType: types.StringType,
			},
		}), diags
	}

	attributeTypes := map[string]attr.Type{
		"link_type": basetypes.StringType{},
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
			"link_type":   v.LinkType,
			"type":        v.ItemsType,
			"validations": validationsVal,
		})

	return objVal, diags
}

func (v ItemsValue) Equal(o attr.Value) bool {
	other, ok := o.(ItemsValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.LinkType.Equal(other.LinkType) {
		return false
	}

	if !v.ItemsType.Equal(other.ItemsType) {
		return false
	}

	if !v.Validations.Equal(other.Validations) {
		return false
	}

	return true
}

func (v ItemsValue) Type(ctx context.Context) attr.Type {
	return ItemsType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v ItemsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"link_type": basetypes.StringType{},
		"type":      basetypes.StringType{},
		"validations": basetypes.ListType{
			ElemType: types.StringType,
		},
	}
}
