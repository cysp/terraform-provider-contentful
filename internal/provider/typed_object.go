package provider

import (
	"context"
	"fmt"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type TypedObject[T any] struct {
	value T
	state attr.ValueState
}

func NewTypedObjectUnknown[T any]() TypedObject[T] {
	return TypedObject[T]{
		state: attr.ValueStateUnknown,
	}
}

func NewTypedObjectNull[T any]() TypedObject[T] {
	return TypedObject[T]{
		state: attr.ValueStateNull,
	}
}

func NewTypedObject[T any](value T) TypedObject[T] {
	return TypedObject[T]{
		value: value,
		state: attr.ValueStateKnown,
	}
}

func NewTypedObjectFromAttributes[T any](ctx context.Context, attributes map[string]attr.Value) (TypedObject[T], diag.Diagnostics) {
	var value T

	setAttributesDiags := tpfr.SetAttributeValues(ctx, &value, attributes)

	return TypedObject[T]{
		value: value,
		state: attr.ValueStateKnown,
	}, setAttributesDiags
}

var _ attr.Value = (*TypedObject[any])(nil)

//nolint:ireturn
func (v TypedObject[T]) Type(ctx context.Context) attr.Type {
	attributeTypes := tpfr.AttributeTypesFor[T](ctx)

	return TypedObjectType[T]{attributeTypes: attributeTypes}
}

func (v TypedObject[T]) CustomType(ctx context.Context) TypedObjectType[T] {
	attributeTypes := tpfr.AttributeTypesFor[T](ctx)

	return TypedObjectType[T]{attributeTypes: attributeTypes}
}

func (v TypedObject[T]) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	tft := v.Type(ctx).TerraformType(ctx)

	if v.IsNull() {
		return tftypes.NewValue(tft, nil), nil
	}

	if v.IsUnknown() {
		return tftypes.NewValue(tft, tftypes.UnknownValue), nil
	}

	attributeValues := tpfr.AttributeValuesOf(v.Value())

	tfval := make(map[string]tftypes.Value, len(attributeValues))

	for key, element := range attributeValues {
		tfvalelem, err := element.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(tft, tftypes.UnknownValue), err
		}

		tfval[key] = tfvalelem
	}

	return tftypes.NewValue(tft, tfval), nil
}

func (v TypedObject[T]) Equal(o attr.Value) bool {
	other, ok := o.(TypedObject[T])
	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	vValue := v.Value()
	otherValue := other.Value()

	attributeTypes := tpfr.AttributeTypesOf(context.Background(), vValue)
	keys := make(map[string]struct{}, len(attributeTypes))

	attributeValues := tpfr.AttributeValuesOf(vValue)
	for k := range attributeValues {
		keys[k] = struct{}{}
	}

	otherAttributeValues := tpfr.AttributeValuesOf(otherValue)
	for k := range otherAttributeValues {
		keys[k] = struct{}{}
	}

	for k := range keys {
		aElementK, aElementKFound := attributeValues[k]
		if !aElementKFound {
			return false
		}

		bElementK, bElementKFound := otherAttributeValues[k]
		if !bElementKFound {
			return false
		}

		if !aElementK.Equal(bElementK) {
			return false
		}
	}

	return true
}

func (v TypedObject[T]) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v TypedObject[T]) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v TypedObject[T]) String() string {
	var t T

	return fmt.Sprintf("TypedObject[%T]", t)
}

var _ basetypes.ObjectValuable = (*TypedObject[struct{}])(nil)

func (v TypedObject[T]) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	attributeTypes := tpfr.AttributeTypesFor[T](ctx)

	if v.IsNull() {
		return types.ObjectNull(attributeTypes), nil
	}

	if v.IsUnknown() {
		return types.ObjectUnknown(attributeTypes), nil
	}

	attributes := tpfr.AttributeValuesOf(v.Value())

	return types.ObjectValue(attributeTypes, attributes)
}

// ---

//nolint:ireturn
func (v TypedObject[T]) Value() T {
	return v.value
}

//nolint:ireturn
func (v TypedObject[T]) GetValue() (T, bool) {
	if v.IsNull() || v.IsUnknown() {
		var zero T

		return zero, false
	}

	return v.value, true
}
