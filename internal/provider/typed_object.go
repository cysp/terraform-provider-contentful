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
func (to TypedObject[T]) Type(ctx context.Context) attr.Type {
	attributeTypes := tpfr.AttributeTypesFor[T](ctx)

	return TypedObjectType[T]{attributeTypes: attributeTypes}
}

func (to TypedObject[T]) CustomType(ctx context.Context) TypedObjectType[T] {
	attributeTypes := tpfr.AttributeTypesFor[T](ctx)

	return TypedObjectType[T]{attributeTypes: attributeTypes}
}

func (to TypedObject[T]) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	tft := to.Type(ctx).TerraformType(ctx)

	if to.IsNull() {
		return tftypes.NewValue(tft, nil), nil
	}

	if to.IsUnknown() {
		return tftypes.NewValue(tft, tftypes.UnknownValue), nil
	}

	attributeValues := tpfr.AttributeValuesOf(to.Value())

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

func (to TypedObject[T]) Equal(other attr.Value) bool {
	otherObject, ok := other.(TypedObject[T])
	if !ok {
		return false
	}

	if to.state != otherObject.state {
		return false
	}

	thisValue := to.Value()
	otherValue := otherObject.Value()

	attributeTypes := tpfr.AttributeTypesOf(context.Background(), thisValue)
	keys := make(map[string]struct{}, len(attributeTypes))

	attributeValues := tpfr.AttributeValuesOf(thisValue)
	for k := range attributeValues {
		keys[k] = struct{}{}
	}

	otherAttributeValues := tpfr.AttributeValuesOf(otherValue)
	for k := range otherAttributeValues {
		keys[k] = struct{}{}
	}

	for k := range keys {
		thisAttribute, thisAttributeFound := attributeValues[k]
		if !thisAttributeFound {
			return false
		}

		otherAttribute, otherAttributeFound := otherAttributeValues[k]
		if !otherAttributeFound {
			return false
		}

		if !thisAttribute.Equal(otherAttribute) {
			return false
		}
	}

	return true
}

func (to TypedObject[T]) IsNull() bool {
	return to.state == attr.ValueStateNull
}

func (to TypedObject[T]) IsUnknown() bool {
	return to.state == attr.ValueStateUnknown
}

func (to TypedObject[T]) String() string {
	var t T

	return fmt.Sprintf("TypedObject[%T]", t)
}

var _ basetypes.ObjectValuable = (*TypedObject[struct{}])(nil)

func (to TypedObject[T]) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	attributeTypes := tpfr.AttributeTypesFor[T](ctx)

	if to.IsNull() {
		return types.ObjectNull(attributeTypes), nil
	}

	if to.IsUnknown() {
		return types.ObjectUnknown(attributeTypes), nil
	}

	attributes := tpfr.AttributeValuesOf(to.Value())

	return types.ObjectValue(attributeTypes, attributes)
}

// ---

//nolint:ireturn
func (to TypedObject[T]) Value() T {
	return to.value
}

//nolint:ireturn
func (to TypedObject[T]) GetValue() (T, bool) {
	if to.IsNull() || to.IsUnknown() {
		var zero T

		return zero, false
	}

	return to.value, true
}
