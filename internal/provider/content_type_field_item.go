package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

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
