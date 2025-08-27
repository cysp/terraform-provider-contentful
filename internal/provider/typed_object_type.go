package provider

import (
	"context"
	"fmt"
	"reflect"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
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
	var v T

	if tfval.Type() == nil {
		return NewTypedObjectNull[T](), nil
	}

	attributeTypes := t.AttributeTypes()

	if !tfval.Type().Equal(t.TerraformType(ctx)) {
		//nolint:err113
		return nil, fmt.Errorf("can't use %s as value of TypedObject[%T]", tfval.String(), v)
	}

	if !tfval.IsKnown() {
		return NewTypedObjectUnknown[T](), nil
	}

	if tfval.IsNull() {
		return NewTypedObjectNull[T](), nil
	}

	tfelems := map[string]tftypes.Value{}

	tfelemsErr := tfval.As(&tfelems)
	if tfelemsErr != nil {
		return nil, fmt.Errorf("error extracting elements from terraform value: %w", tfelemsErr)
	}

	attributeValues := make(map[string]attr.Value, len(tfelems))

	for key, elem := range tfelems {
		attrtyp, attrtypFound := attributeTypes[key]
		if !attrtypFound {
			//nolint:err113
			return nil, fmt.Errorf("unknown attribute %q for TypedObject[%T]", key, v)
		}

		attrval, attrvalErr := attrtyp.ValueFromTerraform(ctx, elem)
		if attrvalErr != nil {
			return nil, fmt.Errorf("error converting element from terraform value: %w", attrvalErr)
		}

		attributeValues[key] = attrval
	}

	typ := reflect.TypeFor[T]()

	val, valOk := reflect.New(typ).Elem().Interface().(T)
	if !valOk {
		//nolint:err113
		return nil, fmt.Errorf("error converting value to type %T", val)
	}

	diags := diag.Diagnostics{}

	setAttributesDiags := tpfr.SetAttributeValues(ctx, &val, attributeValues)
	diags.Append(setAttributesDiags...)

	list := NewTypedObject(val)

	return list, ErrorFromDiags(diags)
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

	for k := range attributeTypes {
		if !otherAttributeTypes[k].Equal(attributeTypes[k]) {
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

//nolint:ireturn
func (t TypedObjectType[T]) WithAttributeTypes(attributeTypes map[string]attr.Type) attr.TypeWithAttributeTypes {
	return TypedObjectType[struct{}]{attributeTypes: attributeTypes}
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
