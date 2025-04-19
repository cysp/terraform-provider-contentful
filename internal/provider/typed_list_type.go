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
func (l TypedListType[T]) TerraformType(ctx context.Context) tftypes.Type {
	var t T

	return tftypes.List{ElementType: t.Type(ctx).TerraformType(ctx)}
}

//nolint:ireturn
func (l TypedListType[T]) ValueFromTerraform(ctx context.Context, tfval tftypes.Value) (attr.Value, error) {
	if tfval.Type() == nil {
		return NewTypedListNull[T](ctx), nil
	}

	elementType := l.ElementTypeWithContext(ctx)

	if !tfval.Type().Equal(l.TerraformType(ctx)) {
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
func (l TypedListType[T]) ValueType(context.Context) attr.Value {
	return TypedList[T]{}
}

func (l TypedListType[T]) Equal(o attr.Type) bool {
	other, ok := o.(TypedListType[T])
	if !ok {
		return false
	}

	elementType := l.ElementType()
	otherElementType := other.ElementType()

	if elementType == nil && otherElementType == nil {
		return true
	}

	return elementType.Equal(otherElementType)
}

func (l TypedListType[T]) String() string {
	var t T

	return fmt.Sprintf("TypedList[%T]", t)
}

func (l TypedListType[T]) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	var t T

	if _, ok := step.(tftypes.ElementKeyInt); !ok {
		//nolint:err113
		return nil, fmt.Errorf("cannot apply step %T to TypedList[%T]", step, t)
	}

	return l.ElementType(), nil
}

//nolint:ireturn
func (l TypedListType[T]) WithElementType(typ attr.Type) attr.TypeWithElementType {
	return TypedListType[attr.Value]{elementType: typ}
}

//nolint:ireturn
func (l TypedListType[T]) ElementType() attr.Type {
	return l.ElementTypeWithContext(context.Background())
}

//nolint:ireturn
func (l TypedListType[T]) ElementTypeWithContext(ctx context.Context) attr.Type {
	if l.elementType != nil {
		return l.elementType
	}

	var t T

	return t.Type(ctx)
}

//nolint:ireturn
func (l TypedListType[T]) ValueFromList(ctx context.Context, value basetypes.ListValue) (basetypes.ListValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	var elements []T

	diags.Append(value.ElementsAs(ctx, &elements, false)...)

	return TypedList[T]{elements: elements, state: attr.ValueStateKnown}, diags
}
