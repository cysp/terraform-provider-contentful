package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type TypedListType[T attr.Value] struct {
	elementType attr.Type
}

var (
	_ attr.Type                = (*TypedListType[attr.Value])(nil)
	_ attr.TypeWithElementType = (*TypedListType[attr.Value])(nil)
	_ basetypes.ListTypable    = (*TypedListType[attr.Value])(nil)
)

//nolint:ireturn
func (t TypedListType[T]) TerraformType(ctx context.Context) tftypes.Type {
	var v T

	return tftypes.List{ElementType: v.Type(ctx).TerraformType(ctx)}
}

//nolint:ireturn
func (t TypedListType[T]) ValueFromTerraform(ctx context.Context, tfval tftypes.Value) (attr.Value, error) {
	if tfval.Type() == nil {
		return NewTypedListNull[T](ctx), nil
	}

	elementType := t.ElementTypeWithContext(ctx)

	if !tfval.Type().Equal(t.TerraformType(ctx)) {
		//nolint:err113
		return nil, fmt.Errorf("can't use %s as value of TypedList[%T], can only use %s values", tfval.String(), elementType, elementType.TerraformType(ctx).String())
	}

	if !tfval.IsKnown() {
		return NewTypedListUnknown[T](ctx), nil
	}

	if tfval.IsNull() {
		return NewTypedListNull[T](ctx), nil
	}

	tfelems := []tftypes.Value{}

	tfelemsErr := tfval.As(&tfelems)
	if tfelemsErr != nil {
		return nil, fmt.Errorf("error extracting elements from terraform value: %w", tfelemsErr)
	}

	elements := make([]T, len(tfelems))

	for idx, elem := range tfelems {
		attrval, attrvalErr := elementType.ValueFromTerraform(ctx, elem)
		if attrvalErr != nil {
			return nil, fmt.Errorf("error converting element from terraform value: %w", attrvalErr)
		}

		element, elementOk := attrval.(T)
		if !elementOk {
			//nolint:err113
			return nil, fmt.Errorf("can't use %s as value of TypedList[%T], can only use %s values", element.String(), elementType, elementType.TerraformType(ctx).String())
		}

		elements[idx] = element
	}

	list, listDiags := NewTypedList(ctx, elements)

	return list, ErrorFromDiags(listDiags)
}

//nolint:ireturn
func (t TypedListType[T]) ValueType(context.Context) attr.Value {
	return TypedList[T]{}
}

func (t TypedListType[T]) Equal(o attr.Type) bool {
	other, ok := o.(TypedListType[T])
	if !ok {
		return false
	}

	elementType := t.ElementType()
	otherElementType := other.ElementType()

	if elementType == nil && otherElementType == nil {
		return true
	}

	return elementType.Equal(otherElementType)
}

func (t TypedListType[T]) String() string {
	var v T

	return fmt.Sprintf("TypedList[%T]", v)
}

func (t TypedListType[T]) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	var v T

	if _, ok := step.(tftypes.ElementKeyInt); !ok {
		//nolint:err113
		return nil, fmt.Errorf("cannot apply step %T to TypedList[%T]", step, v)
	}

	return t.ElementType(), nil
}

//nolint:ireturn
func (t TypedListType[T]) WithElementType(typ attr.Type) attr.TypeWithElementType {
	return TypedListType[attr.Value]{elementType: typ}
}

//nolint:ireturn
func (t TypedListType[T]) ElementType() attr.Type {
	return t.ElementTypeWithContext(context.Background())
}

//nolint:ireturn
func (t TypedListType[T]) ElementTypeWithContext(ctx context.Context) attr.Type {
	if t.elementType != nil {
		return t.elementType
	}

	var v T

	return v.Type(ctx)
}

//nolint:ireturn
func (t TypedListType[T]) ValueFromList(ctx context.Context, value basetypes.ListValue) (basetypes.ListValuable, diag.Diagnostics) {
	if value.IsUnknown() {
		return NewTypedListUnknown[T](ctx), nil
	}

	if value.IsNull() {
		return NewTypedListNull[T](ctx), nil
	}

	var diags diag.Diagnostics

	var elements []T

	diags.Append(value.ElementsAs(ctx, &elements, false)...)

	return TypedList[T]{elements: elements, state: attr.ValueStateKnown}, diags
}
