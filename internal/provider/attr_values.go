package provider

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

func ReflectToTerraformValue(ctx context.Context, value attr.Value, state attr.ValueState) (tftypes.Value, error) {
	tft := value.Type(ctx).TerraformType(ctx)

	switch state {
	case attr.ValueStateKnown:
		break
	case attr.ValueStateNull:
		return tftypes.NewValue(tft, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(tft, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", state))
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

	tfval := make(map[string]tftypes.Value, numAttributes)

	errs := make([]error, 0, numAttributes)

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

		tfFieldValue, err := fieldValue.ToTerraformValue(ctx)
		if err != nil {
			errs = append(errs, err)
		}

		tfval[tag] = tfFieldValue
	}

	validateErr := tftypes.ValidateValue(tft, tfval)
	if validateErr != nil {
		errs = append(errs, validateErr)
	}

	err := errors.Join(errs...)
	if err != nil {
		return tftypes.NewValue(tft, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(tft, tfval), nil
}

func ReflectToObjectValue(ctx context.Context, value AttrValueWithObjectAttrTypes) (basetypes.ObjectValue, diag.Diagnostics) {
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

type UnexpectedTerraformTypeError struct {
	Expected tftypes.Type
	Actual   tftypes.Type
}

func (e UnexpectedTerraformTypeError) Error() string {
	return fmt.Sprintf("expected %s, actual %s", e.Expected, e.Actual)
}
