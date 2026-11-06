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
		state: attr.ValueStateUnknown,
	}
}

func NewTypedMapNull[T attr.Value]() TypedMap[T] {
	return TypedMap[T]{
		state: attr.ValueStateNull,
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
	typ := v.CustomType(ctx)
	terraformType := typ.TerraformType(ctx)

	if v.IsNull() {
		return tftypes.NewValue(terraformType, nil), nil
	}

	if v.IsUnknown() {
		return tftypes.NewValue(terraformType, tftypes.UnknownValue), nil
	}

	// Raw encoding must not invoke the custom-value validation in ToMapValue.
	elements := make(map[string]attr.Value, len(v.elements))
	for key, element := range v.elements {
		elements[key] = element
	}

	value, diags := types.MapValue(typ.ElementTypeWithContext(ctx), elements)
	if diags.HasError() {
		return tftypes.NewValue(terraformType, tftypes.UnknownValue), ErrorFromDiags(diags)
	}

	result, err := value.ToTerraformValue(ctx)
	if err != nil {
		return result, fmt.Errorf("converting typed map to Terraform: %w", err)
	}

	return result, nil
}

func (v TypedMap[T]) Equal(o attr.Value) bool {
	other, ok := o.(TypedMap[T])
	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if len(v.elements) != len(other.elements) {
		return false
	}

	for key, element := range v.elements {
		otherElement, ok := other.elements[key]
		if !ok {
			return false
		}

		if !element.Equal(otherElement) {
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

	elements := make(map[string]attr.Value, len(v.elements))
	for key, element := range v.elements {
		elements[key] = element
	}

	// Framework configuration validation relies on this reflection adapter to
	// validate custom elements. The non-nil map preserves known empty values.
	return types.MapValueFrom(ctx, elementType, elements)
}

// ---

func (v TypedMap[T]) Elements() map[string]T {
	return v.elements
}
