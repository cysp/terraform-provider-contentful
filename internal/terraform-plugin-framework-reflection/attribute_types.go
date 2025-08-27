package terraformpluginframeworkreflection

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

func AttributeTypesFor[T any](ctx context.Context) map[string]attr.Type {
	typ := reflect.TypeFor[T]()

	return attributeTypesOf(ctx, typ)
}

func AttributeTypesOf(ctx context.Context, value any) map[string]attr.Type {
	typ := reflect.TypeOf(value)

	return attributeTypesOf(ctx, typ)
}

func attributeTypesOf(ctx context.Context, typ reflect.Type) map[string]attr.Type {
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	attrs := make(map[string]attr.Type)

	for i := range typ.NumField() {
		field := typ.Field(i)

		tag := field.Tag.Get("tfsdk")
		if tag == "" {
			continue
		}

		fieldValueInterface := reflect.New(field.Type).Interface()

		fieldAttrValue, fieldAttrValueOk := fieldValueInterface.(attr.Value)
		if !fieldAttrValueOk {
			continue
		}

		attrs[tag] = fieldAttrValue.Type(ctx)
	}

	return attrs
}
