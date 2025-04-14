//nolint:dupl
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type EditorInterfaceEditorLayoutElementItemFieldType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = EditorInterfaceEditorLayoutElementItemFieldType{}

func (t EditorInterfaceEditorLayoutElementItemFieldType) Equal(o attr.Type) bool {
	other, ok := o.(EditorInterfaceEditorLayoutElementItemFieldType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t EditorInterfaceEditorLayoutElementItemFieldType) String() string {
	return "EditorInterfaceEditorLayoutElementItemFieldType"
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutElementItemFieldType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, EditorInterfaceEditorLayoutElementItemFieldValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutElementItemFieldType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewEditorInterfaceEditorLayoutElementItemFieldValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewEditorInterfaceEditorLayoutElementItemFieldValueNull(), nil
	}

	if !value.IsKnown() {
		return NewEditorInterfaceEditorLayoutElementItemFieldValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create EditorInterfaceEditorLayoutElementItemFieldValue from Terraform: %w", err)
	}

	v, diags := NewEditorInterfaceEditorLayoutElementItemFieldValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutElementItemFieldType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewEditorInterfaceEditorLayoutElementItemFieldValueNull(), nil
	case value.IsUnknown():
		return NewEditorInterfaceEditorLayoutElementItemFieldValueUnknown(), nil
	}

	return NewEditorInterfaceEditorLayoutElementItemFieldValueKnownFromAttributes(ctx, value.Attributes())
}
