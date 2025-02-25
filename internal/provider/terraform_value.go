package provider

import (
	"context"
	"errors"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func ReflectIntoTerraformValue(ctx context.Context, value basetypes.ObjectValuable) (tftypes.Value, error) {
	tftyp := value.Type(ctx).TerraformType(ctx)

	switch {
	case value.IsNull():
		return tftypes.NewValue(tftyp, nil), nil
	case value.IsUnknown():
		return tftypes.NewValue(tftyp, tftypes.UnknownValue), nil
	}

	valueTyp := reflect.TypeOf(value)
	if valueTyp.Kind() != reflect.Struct {
		return tftypes.NewValue(tftyp, tftypes.UnknownValue), errors.New("expected struct")
	}

	numValueFields := valueTyp.NumField()

	val := make(map[string]tftypes.Value, numValueFields)
	errs := make([]error, 0, numValueFields+1)

	for fieldIndex := range numValueFields {
		field := valueTyp.Field(fieldIndex)

		tag := field.Tag.Get("tfsdk")
		if tag == "" {
			continue
		}

		value := reflect.ValueOf(value).Field(fieldIndex).Interface()

		tfvalue, err := value.(TerraformValuable).ToTerraformValue(ctx)
		if err != nil {
			errs = append(errs, err)
		}

		val[tag] = tfvalue
	}

	validateErr := tftypes.ValidateValue(tftyp, val)
	if validateErr != nil {
		errs = append(errs, validateErr)
	}

	err := errors.Join(errs...)
	if err != nil {
		return tftypes.NewValue(tftyp, tftypes.UnknownValue), err
	}

	return tftypes.NewValue(tftyp, val), nil
}
