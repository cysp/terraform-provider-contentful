package terraformpluginframeworkreflection

import (
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

func AttributeValuesOf(value any) map[string]attr.Value {
	attributeValues := make(map[string]attr.Value)

	extractAttributeValuesOf(attributeValues, value)

	return attributeValues
}

func extractAttributeValuesOf(attributeValues map[string]attr.Value, value any) {
	typ := reflect.TypeOf(value)
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	for i := range typ.NumField() {
		field := typ.Field(i)

		if field.Type.Kind() == reflect.Struct && field.Anonymous {
			fieldValueInterface := val.Field(i).Interface()

			extractAttributeValuesOf(attributeValues, fieldValueInterface)

			continue
		}

		tag := field.Tag.Get("tfsdk")
		if tag == "" {
			continue
		}

		fieldValueInterface := val.Field(i).Interface()

		fieldAttrValue, fieldAttrValueOk := fieldValueInterface.(attr.Value)
		if !fieldAttrValueOk {
			continue
		}

		attributeValues[tag] = fieldAttrValue
	}
}
