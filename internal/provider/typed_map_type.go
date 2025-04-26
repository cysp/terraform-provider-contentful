package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type TypedMapType[T attr.Value] struct {
	elementType attr.Type
}

var (
	_ attr.Type                = (*TypedMapType[attr.Value])(nil)
	_ attr.TypeWithElementType = (*TypedMapType[attr.Value])(nil)
	_ basetypes.MapTypable     = (*TypedMapType[attr.Value])(nil)
)

//nolint:ireturn
func (t TypedMapType[T]) TerraformType(ctx context.Context) tftypes.Type {
	var v T

	return tftypes.Map{ElementType: v.Type(ctx).TerraformType(ctx)}
}

//nolint:ireturn
func (t TypedMapType[T]) ValueFromTerraform(ctx context.Context, tfval tftypes.Value) (attr.Value, error) {
	if tfval.Type() == nil {
		return NewTypedMapNull[T](ctx), nil
	}

	elementType := t.ElementTypeWithContext(ctx)

	if !tfval.Type().Equal(t.TerraformType(ctx)) {
		//nolint:err113
		return nil, fmt.Errorf("can't use %s as value of TypedMap[%T], can only use %s values", tfval.String(), elementType, elementType.TerraformType(ctx).String())
	}

	if !tfval.IsKnown() {
		return NewTypedMapUnknown[T](ctx), nil
	}

	if tfval.IsNull() {
		return NewTypedMapNull[T](ctx), nil
	}

	tfelems := map[string]tftypes.Value{}

	tfelemsErr := tfval.As(&tfelems)
	if tfelemsErr != nil {
		return nil, fmt.Errorf("error extracting elements from terraform value: %w", tfelemsErr)
	}

	elements := make(map[string]T, len(tfelems))

	for key, elem := range tfelems {
		attrval, attrvalErr := elementType.ValueFromTerraform(ctx, elem)
		if attrvalErr != nil {
			return nil, fmt.Errorf("error converting element from terraform value: %w", attrvalErr)
		}

		element, elementOk := attrval.(T)
		if !elementOk {
			//nolint:err113
			return nil, fmt.Errorf("can't use %s as value of TypedMap[%T], can only use %s values", element.String(), elementType, elementType.TerraformType(ctx).String())
		}

		elements[key] = element
	}

	list, listDiags := NewTypedMap(ctx, elements)

	return list, ErrorFromDiags(listDiags)
}

//nolint:ireturn
func (t TypedMapType[T]) ValueType(context.Context) attr.Value {
	return TypedMap[T]{}
}

func (t TypedMapType[T]) Equal(o attr.Type) bool {
	other, ok := o.(TypedMapType[T])
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

func (t TypedMapType[T]) String() string {
	var v T

	return fmt.Sprintf("TypedMap[%T]", v)
}

func (t TypedMapType[T]) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	var v T

	if _, ok := step.(tftypes.ElementKeyString); !ok {
		//nolint:err113
		return nil, fmt.Errorf("cannot apply step %T to TypedMap[%T]", step, v)
	}

	return t.ElementType(), nil
}

//nolint:ireturn
func (t TypedMapType[T]) WithElementType(typ attr.Type) attr.TypeWithElementType {
	return TypedMapType[attr.Value]{elementType: typ}
}

//nolint:ireturn
func (t TypedMapType[T]) ElementType() attr.Type {
	return t.ElementTypeWithContext(context.Background())
}

//nolint:ireturn
func (t TypedMapType[T]) ElementTypeWithContext(ctx context.Context) attr.Type {
	if t.elementType != nil {
		return t.elementType
	}

	var v T

	return v.Type(ctx)
}

//nolint:ireturn
func (t TypedMapType[T]) ValueFromMap(ctx context.Context, value basetypes.MapValue) (basetypes.MapValuable, diag.Diagnostics) {
	if value.IsUnknown() {
		return NewTypedMapUnknown[T](ctx), nil
	}

	if value.IsNull() {
		return NewTypedMapNull[T](ctx), nil
	}

	var diags diag.Diagnostics

	var elements map[string]T

	diags.Append(value.ElementsAs(ctx, &elements, false)...)

	return TypedMap[T]{elements: elements, state: attr.ValueStateKnown}, diags
}
