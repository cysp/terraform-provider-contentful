package terraformpluginframeworkreflection

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func SetAttributeValues(ctx context.Context, value any, attributes map[string]attr.Value) diag.Diagnostics {
	diags := diag.Diagnostics{}

	typ := reflect.TypeOf(value)
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	attributeKeysRemaining := map[string]struct{}{}
	for key := range attributes {
		attributeKeysRemaining[key] = struct{}{}
	}

	for i := range typ.NumField() {
		field := typ.Field(i)

		if field.PkgPath != "" {
			continue
		}

		tag := field.Tag.Get("tfsdk")
		if tag == "" {
			continue
		}

		delete(attributeKeysRemaining, tag)

		fieldValueInterface := reflect.New(field.Type).Interface()

		fieldAttrValue, fieldAttrValueOk := fieldValueInterface.(attr.Value)
		if !fieldAttrValueOk {
			diags.AddAttributeError(path.Root(tag), "invalid data", fmt.Sprintf("expected field of type %s, got %s", "attr.Value", fieldValueInterface))

			continue
		}

		attributeValue, attributeValueFound := attributes[tag]
		if !attributeValueFound || attributeValue == nil {
			diags.AddAttributeError(path.Root(tag), "invalid data", "attribute missing: "+tag)

			continue
		}

		fieldValue := reflect.ValueOf(attributeValue)
		if fieldValue.Kind() == reflect.Pointer {
			fieldValue = fieldValue.Elem()
		}

		if !fieldValue.CanConvert(field.Type) {
			diags.AddAttributeError(path.Root(tag), "invalid data", fmt.Sprintf("expected object of type %s, got %s", fieldAttrValue.Type(ctx), attributeValue.Type(ctx)))

			continue
		}

		val.FieldByIndex(field.Index).Set(fieldValue)
	}

	for key := range attributeKeysRemaining {
		diags.AddAttributeError(path.Root(key), "invalid data", "unknown attribute: "+key)
	}

	return diags
}
