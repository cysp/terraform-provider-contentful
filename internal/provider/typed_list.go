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

func NewTypedListUnknown[T attr.Value](_ context.Context) TypedList[T] {
	return TypedList[T]{
		elements: make([]T, 0),
		state:    attr.ValueStateUnknown,
	}
}

func NewTypedListNull[T attr.Value](_ context.Context) TypedList[T] {
	return TypedList[T]{
		elements: make([]T, 0),
		state:    attr.ValueStateNull,
	}
}

func NewTypedList[T attr.Value](_ context.Context, elements []T) (TypedList[T], diag.Diagnostics) {
	return TypedList[T]{
		elements: elements,
		state:    attr.ValueStateKnown,
	}, nil
}

var _ attr.Value = (*TypedList[attr.Value])(nil)

//nolint:ireturn
func (l TypedList[T]) Type(ctx context.Context) attr.Type {
	var t T

	return TypedListType[T]{elementType: t.Type(ctx)}
}

func (l TypedList[T]) CustomType(ctx context.Context) TypedListType[T] {
	var t T

	return TypedListType[T]{elementType: t.Type(ctx)}
}

func (l TypedList[T]) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	tft := l.Type(ctx).TerraformType(ctx)

	if l.IsNull() {
		return tftypes.NewValue(tft, nil), nil
	}

	if l.IsUnknown() {
		return tftypes.NewValue(tft, tftypes.UnknownValue), nil
	}

	tfval := make([]tftypes.Value, len(l.elements))

	for idx, element := range l.elements {
		tfvalelem, err := element.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(tft, tftypes.UnknownValue), err
		}

		tfval[idx] = tfvalelem
	}

	return tftypes.NewValue(tft, tfval), nil
}

func (l TypedList[T]) Equal(o attr.Value) bool {
	other, ok := o.(TypedList[T])
	if !ok {
		return false
	}

	if l.state != other.state {
		return false
	}

	if l.elements == nil && other.elements == nil {
		return true
	}

	if len(l.elements) != len(other.elements) {
		return false
	}

	for i := range l.elements {
		if !l.elements[i].Equal(other.elements[i]) {
			return false
		}
	}

	return true
}

func (l TypedList[T]) IsNull() bool {
	return l.state == attr.ValueStateNull
}

func (l TypedList[T]) IsUnknown() bool {
	return l.state == attr.ValueStateUnknown
}

func (l TypedList[T]) String() string {
	var t T

	return fmt.Sprintf("TypedList[%T]", t)
}

var _ basetypes.ListValuable = (*TypedList[attr.Value])(nil)

func (l TypedList[T]) ToListValue(ctx context.Context) (basetypes.ListValue, diag.Diagnostics) {
	var t T
	elementType := t.Type(ctx)

	if l.IsNull() {
		return types.ListNull(elementType), nil
	}

	if l.IsUnknown() {
		return types.ListUnknown(elementType), nil
	}

	return types.ListValueFrom(ctx, elementType, l.elements)
}

// ---

func (l TypedList[T]) Elements() []T {
	return l.elements
}
