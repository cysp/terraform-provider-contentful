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
		state: attr.ValueStateUnknown,
	}
}

func NewTypedListNull[T attr.Value]() TypedList[T] {
	return TypedList[T]{
		state: attr.ValueStateNull,
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
func (v TypedList[T]) Type(ctx context.Context) attr.Type {
	var t T

	return TypedListType[T]{elementType: t.Type(ctx)}
}

func (v TypedList[T]) CustomType(ctx context.Context) TypedListType[T] {
	var t T

	return TypedListType[T]{elementType: t.Type(ctx)}
}

func (v TypedList[T]) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	typ := v.CustomType(ctx)
	terraformType := typ.TerraformType(ctx)

	if v.IsNull() {
		return tftypes.NewValue(terraformType, nil), nil
	}

	if v.IsUnknown() {
		return tftypes.NewValue(terraformType, tftypes.UnknownValue), nil
	}

	// Raw encoding must not invoke the custom-value validation in ToListValue.
	elements := make([]attr.Value, len(v.elements))
	for key, element := range v.elements {
		elements[key] = element
	}

	value, diags := types.ListValue(typ.ElementTypeWithContext(ctx), elements)
	if diags.HasError() {
		return tftypes.NewValue(terraformType, tftypes.UnknownValue), ErrorFromDiags(diags)
	}

	result, err := value.ToTerraformValue(ctx)
	if err != nil {
		return result, fmt.Errorf("converting typed list to Terraform: %w", err)
	}

	return result, nil
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

	elements := make([]attr.Value, len(v.elements))
	for key, element := range v.elements {
		elements[key] = element
	}

	// Framework configuration validation relies on this reflection adapter to
	// validate custom elements. The non-nil slice preserves known empty values.
	return types.ListValueFrom(ctx, elementType, elements)
}

// ---

func (v TypedList[T]) Elements() []T {
	return v.elements
}
