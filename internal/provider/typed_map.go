package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type TypedMap[T attr.Value] struct {
	elements map[string]T
	state    attr.ValueState
}

func NewTypedMapUnknown[T attr.Value]() TypedMap[T] {
	return TypedMap[T]{
		elements: make(map[string]T, 0),
		state:    attr.ValueStateUnknown,
	}
}

func NewTypedMapNull[T attr.Value]() TypedMap[T] {
	return TypedMap[T]{
		elements: make(map[string]T, 0),
		state:    attr.ValueStateNull,
	}
}

func NewTypedMap[T attr.Value](elements map[string]T) TypedMap[T] {
	return TypedMap[T]{
		elements: elements,
		state:    attr.ValueStateKnown,
	}
}

var _ attr.Value = (*TypedMap[attr.Value])(nil)

//nolint:ireturn
func (tm TypedMap[T]) Type(ctx context.Context) attr.Type {
	var t T

	return TypedMapType[T]{elementType: t.Type(ctx)}
}

func (tm TypedMap[T]) CustomType(ctx context.Context) TypedMapType[T] {
	var t T

	return TypedMapType[T]{elementType: t.Type(ctx)}
}

func (tm TypedMap[T]) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	tft := tm.Type(ctx).TerraformType(ctx)

	if tm.IsNull() {
		return tftypes.NewValue(tft, nil), nil
	}

	if tm.IsUnknown() {
		return tftypes.NewValue(tft, tftypes.UnknownValue), nil
	}

	tfval := make(map[string]tftypes.Value, len(tm.elements))

	for key, element := range tm.elements {
		tfvalelem, err := element.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(tft, tftypes.UnknownValue), err
		}

		tfval[key] = tfvalelem
	}

	return tftypes.NewValue(tft, tfval), nil
}

func (tm TypedMap[T]) Equal(other attr.Value) bool {
	otherMap, ok := other.(TypedMap[T])
	if !ok {
		return false
	}

	if tm.state != otherMap.state {
		return false
	}

	if tm.elements == nil && otherMap.elements == nil {
		return true
	}

	keys := make(map[string]struct{}, len(tm.elements))
	for k := range tm.elements {
		keys[k] = struct{}{}
	}

	for k := range otherMap.elements {
		keys[k] = struct{}{}
	}

	for k := range keys {
		thisElement, thisElementFound := tm.elements[k]
		if !thisElementFound {
			return false
		}

		otherElement, otherElementFound := otherMap.elements[k]
		if !otherElementFound {
			return false
		}

		if !thisElement.Equal(otherElement) {
			return false
		}
	}

	return true
}

func (tm TypedMap[T]) IsNull() bool {
	return tm.state == attr.ValueStateNull
}

func (tm TypedMap[T]) IsUnknown() bool {
	return tm.state == attr.ValueStateUnknown
}

func (tm TypedMap[T]) String() string {
	var t T

	return fmt.Sprintf("TypedMap[%T]", t)
}

var _ basetypes.MapValuable = (*TypedMap[attr.Value])(nil)

func (tm TypedMap[T]) ToMapValue(ctx context.Context) (basetypes.MapValue, diag.Diagnostics) {
	var t T

	elementType := t.Type(ctx)

	if tm.IsNull() {
		return types.MapNull(elementType), nil
	}

	if tm.IsUnknown() {
		return types.MapUnknown(elementType), nil
	}

	return types.MapValueFrom(ctx, elementType, tm.elements)
}

// ---

func (tm TypedMap[T]) Elements() map[string]T {
	return tm.elements
}

func (tm TypedMap[T]) Has(key string) bool {
	_, exists := tm.elements[key]

	return exists
}

func (tm TypedMap[T]) Set(key string, value T) {
	tm.elements[key] = value
}
