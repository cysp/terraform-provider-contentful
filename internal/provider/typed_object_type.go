package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type TypedObjectType[T any] struct {
	attributeTypes map[string]attr.Type
}

var (
	_ attr.Type                   = (*TypedObjectType[any])(nil)
	_ attr.TypeWithAttributeTypes = (*TypedObjectType[any])(nil)
)

//nolint:ireturn
func (t TypedObjectType[T]) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, t.attributeTypes)}
}

//nolint:ireturn
func (t TypedObjectType[T]) ValueFromTerraform(ctx context.Context, tfval tftypes.Value) (attr.Value, error) {
	value, err := (basetypes.ObjectType{AttrTypes: t.attributeTypes}).ValueFromTerraform(ctx, tfval)
	if err != nil {
		return nil, fmt.Errorf("converting Terraform to typed object: %w", err)
	}

	// The Framework base type always returns its corresponding value type.
	base, ok := value.(basetypes.ObjectValue)
	if !ok {
		//nolint:err113
		return nil, fmt.Errorf("unexpected Framework object value %T", value)
	}

	typed, diags := t.ValueFromObject(ctx, base)

	return typed, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t TypedObjectType[T]) ValueType(context.Context) attr.Value {
	return TypedObject[T]{}
}

func (t TypedObjectType[T]) Equal(o attr.Type) bool {
	other, ok := o.(TypedObjectType[T])
	if !ok {
		return false
	}

	attributeTypes := t.AttributeTypes()
	otherAttributeTypes := other.AttributeTypes()

	if len(attributeTypes) != len(otherAttributeTypes) {
		return false
	}

	for key, attributeType := range attributeTypes {
		otherAttributeType, ok := otherAttributeTypes[key]
		if !ok {
			return false
		}

		if !attributeType.Equal(otherAttributeType) {
			return false
		}
	}

	return true
}

func (t TypedObjectType[T]) String() string {
	var v T

	return fmt.Sprintf("TypedObject[%T]", v)
}

func (t TypedObjectType[T]) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	var v T

	key, keyOk := step.(tftypes.AttributeName)
	if !keyOk {
		//nolint:err113
		return nil, fmt.Errorf("cannot apply step %T to TypedObject[%T]", step, v)
	}

	return t.attributeTypes[string(key)], nil
}

// WithAttributeTypes retains the typed value when the replacement equals the
// descriptor derived from T. Other replacements use the standard Framework type
// so their values can preserve the requested descriptor.
//
//nolint:ireturn
func (t TypedObjectType[T]) WithAttributeTypes(attributeTypes map[string]attr.Type) attr.TypeWithAttributeTypes {
	replacement := TypedObjectType[T]{attributeTypes: attributeTypes}
	canonical := TypedObjectType[T]{attributeTypes: tpfr.AttributeTypesFor[T](context.Background())}

	if replacement.Equal(canonical) {
		return replacement
	}

	return basetypes.ObjectType{AttrTypes: attributeTypes}
}

func (t TypedObjectType[T]) AttributeTypes() map[string]attr.Type {
	return t.attributeTypes
}

var _ basetypes.ObjectTypable = (*TypedObjectType[any])(nil)

//nolint:ireturn
func (t TypedObjectType[T]) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	if value.IsUnknown() {
		return NewTypedObjectUnknown[T](), nil
	}

	if value.IsNull() {
		return NewTypedObjectNull[T](), nil
	}

	var diags diag.Diagnostics

	objectValue, valueDiags := NewTypedObjectFromAttributes[T](ctx, value.Attributes())
	diags.Append(valueDiags...)

	return objectValue, diags
}
