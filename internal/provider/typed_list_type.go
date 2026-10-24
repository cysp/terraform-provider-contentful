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
	value, err := (basetypes.ListType{ElemType: t.ElementTypeWithContext(ctx)}).ValueFromTerraform(ctx, tfval)
	if err != nil {
		return nil, fmt.Errorf("converting Terraform to typed list: %w", err)
	}

	// The Framework base type always returns its corresponding value type.
	base, ok := value.(basetypes.ListValue)
	if !ok {
		//nolint:err113
		return nil, fmt.Errorf("unexpected Framework list value %T", value)
	}

	if base.IsNull() {
		return NewTypedListNull[T](), nil
	}

	if base.IsUnknown() {
		return NewTypedListUnknown[T](), nil
	}

	// ValueFromList uses reflection and validates elements; raw decoding must not.
	elements := base.Elements()
	typed := make([]T, len(elements))

	for key, element := range elements {
		value, ok := element.(T)
		if !ok {
			//nolint:err113
			return nil, fmt.Errorf("unexpected typed list element %T at %v", element, key)
		}

		typed[key] = value
	}

	return NewTypedList(typed), nil
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

// WithElementType retains the typed value when the replacement equals the
// descriptor derived from T. Other replacements use the standard Framework type
// so their values can preserve the requested descriptor.
//
//nolint:ireturn
func (t TypedListType[T]) WithElementType(typ attr.Type) attr.TypeWithElementType {
	var value T
	if typ != nil && typ.Equal(value.Type(context.Background())) {
		return TypedListType[T]{elementType: typ}
	}

	return basetypes.ListType{ElemType: typ}
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
		return NewTypedListUnknown[T](), nil
	}

	if value.IsNull() {
		return NewTypedListNull[T](), nil
	}

	var diags diag.Diagnostics

	var elements []T

	diags.Append(value.ElementsAs(ctx, &elements, false)...)

	return TypedList[T]{elements: elements, state: attr.ValueStateKnown}, diags
}
