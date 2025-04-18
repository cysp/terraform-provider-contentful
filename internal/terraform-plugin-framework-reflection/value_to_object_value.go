package terraformpluginframeworkreflection

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ValueToObjectValue(ctx context.Context, value AttrValueWithObjectAttrTypes) (types.Object, diag.Diagnostics) {
	attributeTypes := value.ObjectAttrTypes(ctx)

	switch {
	case value.IsNull():
		return types.ObjectNull(attributeTypes), nil
	case value.IsUnknown():
		return types.ObjectUnknown(attributeTypes), nil
	}

	typ := reflect.TypeOf(value)
	val := reflect.ValueOf(value)

	numAttributes := 0

	for i := range typ.NumField() {
		field := typ.Field(i)
		if field.Tag.Get("tfsdk") != "" {
			numAttributes++
		}
	}

	attributes := map[string]attr.Value{}

	for i := range typ.NumField() {
		field := typ.Field(i)

		tag := field.Tag.Get("tfsdk")
		if tag == "" {
			continue
		}

		fieldValueInterface := val.FieldByIndex(field.Index).Interface()

		fieldValue, fieldValueOk := fieldValueInterface.(attr.Value)
		if !fieldValueOk {
			continue
		}

		attributes[tag] = fieldValue
	}

	return types.ObjectValue(attributeTypes, attributes)
}
