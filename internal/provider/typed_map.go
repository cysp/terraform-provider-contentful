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
func (v TypedMap[T]) Type(ctx context.Context) attr.Type {
	var t T

	return TypedMapType[T]{elementType: t.Type(ctx)}
}

func (v TypedMap[T]) CustomType(ctx context.Context) TypedMapType[T] {
	var t T

	return TypedMapType[T]{elementType: t.Type(ctx)}
}

func (v TypedMap[T]) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	tft := v.Type(ctx).TerraformType(ctx)

	if v.IsNull() {
		return tftypes.NewValue(tft, nil), nil
	}

	if v.IsUnknown() {
		return tftypes.NewValue(tft, tftypes.UnknownValue), nil
	}

	tfval := make(map[string]tftypes.Value, len(v.elements))

	for key, element := range v.elements {
		tfvalelem, err := element.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(tft, tftypes.UnknownValue), err
		}

		tfval[key] = tfvalelem
	}

	return tftypes.NewValue(tft, tfval), nil
}

func (v TypedMap[T]) Equal(o attr.Value) bool {
	other, ok := o.(TypedMap[T])
	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.elements == nil && other.elements == nil {
		return true
	}

	keys := make(map[string]struct{}, len(v.elements))
	for k := range v.elements {
		keys[k] = struct{}{}
	}

	for k := range other.elements {
		keys[k] = struct{}{}
	}

	for k := range keys {
		aElementK, aElementKFound := v.elements[k]
		if !aElementKFound {
			return false
		}

		bElementK, bElementKFound := other.elements[k]
		if !bElementKFound {
			return false
		}

		if !aElementK.Equal(bElementK) {
			return false
		}
	}

	return true
}

func (v TypedMap[T]) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v TypedMap[T]) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v TypedMap[T]) String() string {
	var t T

	return fmt.Sprintf("TypedMap[%T]", t)
}

var _ basetypes.MapValuable = (*TypedMap[attr.Value])(nil)

func (v TypedMap[T]) ToMapValue(ctx context.Context) (basetypes.MapValue, diag.Diagnostics) {
	var t T

	elementType := t.Type(ctx)

	if v.IsNull() {
		return types.MapNull(elementType), nil
	}

	if v.IsUnknown() {
		return types.MapUnknown(elementType), nil
	}

	return types.MapValueFrom(ctx, elementType, v.elements)
}

// ---

func (v TypedMap[T]) Elements() map[string]T {
	return v.elements
}
