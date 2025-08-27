package terraformpluginframeworkreflection

import (
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

func AttributeValuesOf(value any) map[string]attr.Value {
	typ := reflect.TypeOf(value)
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	attrs := make(map[string]attr.Value)

	for i := range typ.NumField() {
		field := typ.Field(i)

		tag := field.Tag.Get("tfsdk")
		if tag == "" {
			continue
		}

		fieldValueInterface := val.Field(i).Interface()

		fieldAttrValue, fieldAttrValueOk := fieldValueInterface.(attr.Value)
		if !fieldAttrValueOk {
			continue
		}

		attrs[tag] = fieldAttrValue
	}

	return attrs
}
