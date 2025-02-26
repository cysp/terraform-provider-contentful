package provider

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type AttrTypeable interface {
	Type(_ context.Context) attr.Type
}

type AttrValueWithObjectAttrTypes interface {
	attr.Value

	ObjectAttrTypes(ctx context.Context) map[string]attr.Type
}

type AttrTypeWithValueFromObject interface {
	attr.Type

	ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics)
}

type AttrValueWithToObjectValue interface {
	attr.Value

	ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics)
}

func ReflectObjectAttrTypes(ctx context.Context, value basetypes.ObjectValuable) map[string]attr.Type {
	valueTyp := reflect.TypeOf(value)
	if valueTyp.Kind() != reflect.Struct {
		panic("expected struct")
	}

	numValueFields := valueTyp.NumField()

	attrTypes := make(map[string]attr.Type, numValueFields)

	for fieldIndex := range numValueFields {
		field := valueTyp.Field(fieldIndex)

		tag := field.Tag.Get("tfsdk")
		if tag == "" {
			continue
		}

		fieldTypeAttrTypeable := field.Type.Implements(reflect.TypeOf((*AttrTypeable)(nil)).Elem())
		if !fieldTypeAttrTypeable {
			continue
		}

		fieldTypeType, fieldTypeTypeOk := field.Type.MethodByName("Type")
		if !fieldTypeTypeOk {
			continue
		}

		fieldType, fieldTypeOk := fieldTypeType.Func.Call([]reflect.Value{reflect.ValueOf(value).Field(fieldIndex), reflect.ValueOf(ctx)})[0].Interface().(attr.Type)
		if !fieldTypeOk {
			continue
		}

		attrTypes[tag] = fieldType
	}

	return attrTypes
}

func ReflectIntoObjectValue(ctx context.Context, value AttrValueWithObjectAttrTypes) (basetypes.ObjectValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	attributeTypes := value.ObjectAttrTypes(ctx)

	switch {
	case value.IsNull():
		return types.ObjectNull(attributeTypes), nil
	case value.IsUnknown():
		return types.ObjectUnknown(attributeTypes), nil
	}

	valueTyp := reflect.TypeOf(value)
	if valueTyp.Kind() != reflect.Struct {
		diags.AddError("expected struct", "")

		return types.ObjectUnknown(attributeTypes), diags
	}

	numValueFields := valueTyp.NumField()

	attributes := make(map[string]attr.Value, numValueFields)

	for fieldIndex := range numValueFields {
		field := valueTyp.Field(fieldIndex)

		tag := field.Tag.Get("tfsdk")
		if tag == "" {
			continue
		}

		fieldValue, fieldValueOk := reflect.ValueOf(value).Field(fieldIndex).Interface().(attr.Value)
		if !fieldValueOk {
			diags.AddAttributeError(path.Root(tag), "expected attr.Value", "")
		}

		attributes[tag] = fieldValue
	}

	return types.ObjectValue(attributeTypes, attributes)
}

//nolint:varnamelen
func ReflectAttrValueEqual(a attr.Value, b attr.Value) bool {
	aTyp := reflect.TypeOf(a)
	bTyp := reflect.TypeOf(b)

	if aTyp != bTyp {
		return false
	}

	if a.IsUnknown() != b.IsUnknown() || a.IsNull() != b.IsNull() {
		return false
	}

	if a.IsUnknown() || a.IsNull() {
		return true
	}

	numValueFields := aTyp.NumField()

	aValue := reflect.ValueOf(a)
	bValue := reflect.ValueOf(b)

	for fieldIndex := range numValueFields {
		aFieldValue := aValue.Field(fieldIndex)
		bFieldValue := bValue.Field(fieldIndex)

		aFieldAttrValue, aFieldAttrValueOk := aFieldValue.Interface().(attr.Value)
		bFieldAttrValue, bFieldAttrValueOk := bFieldValue.Interface().(attr.Value)

		if !aFieldAttrValueOk || !bFieldAttrValueOk {
			return false
		}

		if !aFieldAttrValue.Equal(bFieldAttrValue) {
			return false
		}
	}

	return true
}
