package terraformpluginframeworkreflection

import (
	"context"
	"errors"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func ValueToTerraformValue(ctx context.Context, value attr.Value, state attr.ValueState) (tftypes.Value, error) {
	tft := value.Type(ctx).TerraformType(ctx)

	switch state {
	case attr.ValueStateKnown:
		break
	case attr.ValueStateNull:
		return tftypes.NewValue(tft, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(tft, tftypes.UnknownValue), nil
	default:
		return tftypes.NewValue(tft, nil), UnexpectedValueStateError{ValueState: state}
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
