package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type EditorInterfaceEditorLayoutType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = EditorInterfaceEditorLayoutType{}

func (t EditorInterfaceEditorLayoutType) Equal(o attr.Type) bool {
	other, ok := o.(EditorInterfaceEditorLayoutType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutType) ValueType(_ context.Context) attr.Value {
	return EditorInterfaceEditorLayoutValue{}
}

func (t EditorInterfaceEditorLayoutType) String() string {
	return "EditorInterfaceEditorLayoutType"
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, EditorInterfaceEditorLayoutValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewEditorInterfaceEditorLayoutValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewEditorInterfaceEditorLayoutValueNull(), nil
	}

	if !value.IsKnown() {
		return NewEditorInterfaceEditorLayoutValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create EditorInterfaceEditorLayoutValue from Terraform: %w", err)
	}

	v, diags := NewEditorInterfaceEditorLayoutValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewEditorInterfaceEditorLayoutValueNull(), nil
	case value.IsUnknown():
		return NewEditorInterfaceEditorLayoutValueUnknown(), nil
	}

	return NewEditorInterfaceEditorLayoutValueKnownFromAttributes(ctx, value.Attributes())
}
