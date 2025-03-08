package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

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
