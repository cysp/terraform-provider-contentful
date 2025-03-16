package terraformpluginframeworkreflection

import (
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

func ValueAttributesEqual[T attr.Value](a, b T) bool {
	typ := reflect.TypeFor[T]()

	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	equal := true

	for i := range typ.NumField() {
		field := typ.Field(i)

		tag := field.Tag.Get("tfsdk")
		if tag == "" {
			continue
		}

		fieldTypeInterface := reflect.New(field.Type).Interface()

		if _, fieldTypeValueOk := fieldTypeInterface.(attr.Value); fieldTypeValueOk {
			aFieldVal := aVal.FieldByIndex(field.Index)
			bFieldVal := bVal.FieldByIndex(field.Index)

			aFieldValValue, aFieldValValueOk := aFieldVal.Interface().(attr.Value)
			bFieldValValue, bFieldValValueOk := bFieldVal.Interface().(attr.Value)

			if aFieldValValueOk && bFieldValValueOk {
				if !aFieldValValue.Equal(bFieldValValue) {
					equal = false
				}
			}

			continue
		}
	}

	return equal
}
