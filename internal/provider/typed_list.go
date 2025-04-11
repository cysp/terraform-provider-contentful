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
func (v TypedList[T]) Type(ctx context.Context) attr.Type {
	var t T

	return TypedListType[T]{elementType: t.Type(ctx)}
}

func (v TypedList[T]) CustomType(ctx context.Context) TypedListType[T] {
	var t T

	return TypedListType[T]{elementType: t.Type(ctx)}
}

func (v TypedList[T]) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	tft := v.Type(ctx).TerraformType(ctx)

	if v.IsNull() {
		return tftypes.NewValue(tft, nil), nil
	}

	if v.IsUnknown() {
		return tftypes.NewValue(tft, tftypes.UnknownValue), nil
	}

	tfval := make([]tftypes.Value, len(v.elements))

	for idx, element := range v.elements {
		tfvalelem, err := element.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(tft, tftypes.UnknownValue), err
		}

		tfval[idx] = tfvalelem
	}

	return tftypes.NewValue(tft, tfval), nil
}

func (v TypedList[T]) Equal(o attr.Value) bool {
	other, ok := o.(TypedList[T])
	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.elements == nil && other.elements == nil {
		return true
	}

	if len(v.elements) != len(other.elements) {
		return false
	}

	for i := range v.elements {
		if !v.elements[i].Equal(other.elements[i]) {
			return false
		}
	}

	return true
}

func (v TypedList[T]) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v TypedList[T]) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v TypedList[T]) String() string {
	var t T

	return fmt.Sprintf("TypedList[%T]", t)
}

var _ basetypes.ListValuable = (*TypedList[attr.Value])(nil)

func (v TypedList[T]) ToListValue(ctx context.Context) (basetypes.ListValue, diag.Diagnostics) {
	var t T
	elementType := t.Type(ctx)

	if v.IsNull() {
		return types.ListNull(elementType), nil
	}

	if v.IsUnknown() {
		return types.ListUnknown(elementType), nil
	}

	return types.ListValueFrom(ctx, elementType, v.elements)
}

// ---

func (v TypedList[T]) Elements() []T {
	return v.elements
}
