package terraformpluginframeworkreflection

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

func AttributeTypesFor[T any](ctx context.Context) map[string]attr.Type {
	typ := reflect.TypeFor[T]()

	attributeTypes := make(map[string]attr.Type)

	extractAttributeTypesOf(ctx, attributeTypes, typ)

	return attributeTypes
}

func AttributeTypesOf(ctx context.Context, value any) map[string]attr.Type {
	typ := reflect.TypeOf(value)

	attributeTypes := make(map[string]attr.Type)

	extractAttributeTypesOf(ctx, attributeTypes, typ)

	return attributeTypes
}

func extractAttributeTypesOf(ctx context.Context, attributeTypes map[string]attr.Type, typ reflect.Type) {
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	for i := range typ.NumField() {
		field := typ.Field(i)

		if field.Type.Kind() == reflect.Struct && field.Anonymous {
			extractAttributeTypesOf(ctx, attributeTypes, field.Type)

			continue
		}

		tag := field.Tag.Get("tfsdk")
		if tag == "" {
			continue
		}

		fieldValueInterface := reflect.New(field.Type).Interface()

		fieldAttrValue, fieldAttrValueOk := fieldValueInterface.(attr.Value)
		if !fieldAttrValueOk {
			continue
		}

		attributeTypes[tag] = fieldAttrValue.Type(ctx)
	}
}
