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
	value, err := (basetypes.MapType{ElemType: t.ElementTypeWithContext(ctx)}).ValueFromTerraform(ctx, tfval)
	if err != nil {
		return nil, fmt.Errorf("converting Terraform to typed map: %w", err)
	}

	// The Framework base type always returns its corresponding value type.
	base, ok := value.(basetypes.MapValue)
	if !ok {
		//nolint:err113
		return nil, fmt.Errorf("unexpected Framework map value %T", value)
	}

	if base.IsNull() {
		return NewTypedMapNull[T](), nil
	}

	if base.IsUnknown() {
		return NewTypedMapUnknown[T](), nil
	}

	// ValueFromMap uses reflection and validates elements; raw decoding must not.
	elements := base.Elements()
	typed := make(map[string]T, len(elements))

	for key, element := range elements {
		value, ok := element.(T)
		if !ok {
			//nolint:err113
			return nil, fmt.Errorf("unexpected typed map element %T at %v", element, key)
		}

		typed[key] = value
	}

	return NewTypedMap(typed), nil
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
		return NewTypedMapUnknown[T](), nil
	}

	if value.IsNull() {
		return NewTypedMapNull[T](), nil
	}

	var diags diag.Diagnostics

	var elements map[string]T

	diags.Append(value.ElementsAs(ctx, &elements, false)...)

	return TypedMap[T]{elements: elements, state: attr.ValueStateKnown}, diags
}
