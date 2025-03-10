package provider

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func NewEmptyListMust(elementType attr.Type) basetypes.ListValue {
	list, _ := types.ListValue(elementType, []attr.Value{})

	return list
}

func NewEmptySetMust(elementType attr.Type) basetypes.SetValue {
	list, _ := types.SetValue(elementType, []attr.Value{})

	return list
}

func setTFSDKAttributesInValue(ctx context.Context, value attr.Value, attributes map[string]attr.Value) diag.Diagnostics {
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

func compareTFSDKAttributesEqual[T attr.Value](a, b T) bool {
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

func AttributesFromTerraformValue(ctx context.Context, attrTypes map[string]attr.Type, value tftypes.Value) (map[string]attr.Value, error) {
	attributes := map[string]attr.Value{}

	tfvals := map[string]tftypes.Value{}

	err := value.As(&tfvals)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	for key, tfval := range tfvals {
		a, err := attrTypes[key].ValueFromTerraform(ctx, tfval)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		attributes[key] = a
	}

	return attributes, nil
}

type UnexpectedTerraformTypeError struct {
	Expected tftypes.Type
	Actual   tftypes.Type
}

func (e UnexpectedTerraformTypeError) Error() string {
	return fmt.Sprintf("expected %s, actual %s", e.Expected, e.Actual)
}
