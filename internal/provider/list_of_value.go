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

type ListOfType[T attr.Value] struct {
	ElementType attr.Type
	attr.Type
}

// var _ attr.Type = (*ListOfType[any])(nil)
// var _ basetypes.ListTypable = (*ListOfType[any])(nil)

// ApplyTerraform5AttributePathStep implements basetypes.ListTypable.
// Subtle: this method shadows the method (Type).ApplyTerraform5AttributePathStep of ListOfType.Type.
func (l ListOfType[T]) ApplyTerraform5AttributePathStep(tftypes.AttributePathStep) (interface{}, error) {
	panic("unimplemented")
}

// Equal implements basetypes.ListTypable.
// Subtle: this method shadows the method (Type).Equal of ListOfType.Type.
func (l ListOfType[T]) Equal(attr.Type) bool {
	panic("unimplemented")
}

// String implements basetypes.ListTypable.
// Subtle: this method shadows the method (Type).String of ListOfType.Type.
func (l ListOfType[T]) String() string {
	panic("unimplemented")
}

// TerraformType implements basetypes.ListTypable.
// Subtle: this method shadows the method (Type).TerraformType of ListOfType.Type.
func (l ListOfType[T]) TerraformType(context.Context) tftypes.Type {
	panic("unimplemented")
}

// ValueFromList implements basetypes.ListTypable.
func (l ListOfType[T]) ValueFromList(context.Context, basetypes.ListValue) (basetypes.ListValuable, diag.Diagnostics) {
	panic("unimplemented")
}

// ValueFromTerraform implements basetypes.ListTypable.
// Subtle: this method shadows the method (Type).ValueFromTerraform of ListOfType.Type.
func (l ListOfType[T]) ValueFromTerraform(context.Context, tftypes.Value) (attr.Value, error) {
	panic("unimplemented")
}

// ValueType implements basetypes.ListTypable.
// Subtle: this method shadows the method (Type).ValueType of ListOfType.Type.
func (l ListOfType[T]) ValueType(context.Context) attr.Value {
	panic("unimplemented")
}

type ListOf[T attr.Value] struct {
	elements []T
	state  attr.ValueState
}

// Equal implements basetypes.ListValuable.
func (l ListOf[T]) Equal(attr.Value) bool {
	panic("unimplemented")
}

// IsNull implements basetypes.ListValuable.
func (l ListOf[T]) IsNull() bool {
	return l.state == attr.ValueStateNull
}

// IsUnknown implements basetases.ListValuable.
func (l ListOf[T]) IsUnknown() bool {
	return l.state == attr.ValueStateUnknown
}

// String implements basetypes.ListValuable.
func (l ListOf[T]) String() string {
	var t T
	return fmt.Sprintf("ListOf[%T]", t)
}

// ToListValue implements basetypes.ListValuable.
func (l ListOf[T]) ToListValue(ctx context.Context) (basetypes.ListValue, diag.Diagnostics) {
	var t T

	return types.ListValueFrom(ctx, t.Type(ctx), l.elements)
}

// ToTerraformValue implements basetypes.ListValuable.
func (l ListOf[T]) ToTerraformValue(context.Context) (tftypes.Value, error) {
	panic("unimplemented")
}

// Type implements basetypes.ListValuable.
func (l ListOf[T]) Type(ctx context.Context) attr.Type {
	var t T
	v := t.Type(ctx)
	return ListOfType[attr.Value]{ElementType: v}
}

func (l ListOf[T]) Elements() []T {
	return l.elements
}

// var _ basetypes.ListValuable = ListOf[any]{}
