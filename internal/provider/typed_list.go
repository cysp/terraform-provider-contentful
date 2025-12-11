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

type TypedList[T attr.Value] struct {
	elements []T
	state    attr.ValueState
}

func NewTypedListUnknown[T attr.Value]() TypedList[T] {
	return TypedList[T]{
		elements: make([]T, 0),
		state:    attr.ValueStateUnknown,
	}
}

func NewTypedListNull[T attr.Value]() TypedList[T] {
	return TypedList[T]{
		elements: make([]T, 0),
		state:    attr.ValueStateNull,
	}
}

func NewTypedList[T attr.Value](elements []T) TypedList[T] {
	return TypedList[T]{
		elements: elements,
		state:    attr.ValueStateKnown,
	}
}

var _ attr.Value = (*TypedList[attr.Value])(nil)

//nolint:ireturn
func (tl TypedList[T]) Type(ctx context.Context) attr.Type {
	var t T

	return TypedListType[T]{elementType: t.Type(ctx)}
}

func (tl TypedList[T]) CustomType(ctx context.Context) TypedListType[T] {
	var t T

	return TypedListType[T]{elementType: t.Type(ctx)}
}

func (tl TypedList[T]) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	tft := tl.Type(ctx).TerraformType(ctx)

	if tl.IsNull() {
		return tftypes.NewValue(tft, nil), nil
	}

	if tl.IsUnknown() {
		return tftypes.NewValue(tft, tftypes.UnknownValue), nil
	}

	tfval := make([]tftypes.Value, len(tl.elements))

	for idx, element := range tl.elements {
		tfvalelem, err := element.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(tft, tftypes.UnknownValue), err
		}

		tfval[idx] = tfvalelem
	}

	return tftypes.NewValue(tft, tfval), nil
}

func (tl TypedList[T]) Equal(other attr.Value) bool {
	otherList, ok := other.(TypedList[T])
	if !ok {
		return false
	}

	if tl.state != otherList.state {
		return false
	}

	if tl.elements == nil && otherList.elements == nil {
		return true
	}

	if len(tl.elements) != len(otherList.elements) {
		return false
	}

	for i := range tl.elements {
		if !tl.elements[i].Equal(otherList.elements[i]) {
			return false
		}
	}

	return true
}

func (tl TypedList[T]) IsNull() bool {
	return tl.state == attr.ValueStateNull
}

func (tl TypedList[T]) IsUnknown() bool {
	return tl.state == attr.ValueStateUnknown
}

func (tl TypedList[T]) String() string {
	var t T

	return fmt.Sprintf("TypedList[%T]", t)
}

var _ basetypes.ListValuable = (*TypedList[attr.Value])(nil)

func (tl TypedList[T]) ToListValue(ctx context.Context) (basetypes.ListValue, diag.Diagnostics) {
	var t T

	elementType := t.Type(ctx)

	if tl.IsNull() {
		return types.ListNull(elementType), nil
	}

	if tl.IsUnknown() {
		return types.ListUnknown(elementType), nil
	}

	return types.ListValueFrom(ctx, elementType, tl.elements)
}

// ---

func (tl TypedList[T]) Elements() []T {
	return tl.elements
}
