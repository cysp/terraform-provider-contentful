package terraformpluginframeworkreflection

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func SetAttributesInValue(ctx context.Context, value attr.Value, attributes map[string]attr.Value) diag.Diagnostics {
	diags := diag.Diagnostics{}

	typ := reflect.TypeOf(value).Elem()
	val := reflect.ValueOf(value).Elem()

	for i := range typ.NumField() {
		field := typ.Field(i)

		tag := field.Tag.Get("tfsdk")
		if tag == "" {
			continue
		}

		fieldTypeInterface := reflect.New(field.Type).Interface()

		fieldTypeValue, fieldTypeValueOk := fieldTypeInterface.(attr.Value)
		if !fieldTypeValueOk {
			continue
		}

		attributeValue, attributeValueFound := attributes[tag]
		if !attributeValueFound || attributeValue == nil {
			diags.AddAttributeError(path.Root(tag), "attribute missing", "attribute")

			continue
		}

		if !reflect.ValueOf(attributeValue).CanConvert(field.Type) {
			diags.AddAttributeError(path.Root(tag), "invalid data", fmt.Sprintf("expected object of type %s, got %s", fieldTypeValue.Type(ctx), attributeValue.Type(ctx)))

			continue
		}

		val.FieldByIndex(field.Index).Set(reflect.ValueOf(attributeValue))
	}

	return diags
}
