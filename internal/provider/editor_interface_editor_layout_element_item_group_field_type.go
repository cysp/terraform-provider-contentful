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

type EditorInterfaceEditorLayoutElementItemGroupFieldType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = EditorInterfaceEditorLayoutElementItemGroupFieldType{}

func (t EditorInterfaceEditorLayoutElementItemGroupFieldType) Equal(o attr.Type) bool {
	other, ok := o.(EditorInterfaceEditorLayoutElementItemGroupFieldType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t EditorInterfaceEditorLayoutElementItemGroupFieldType) String() string {
	return "EditorInterfaceEditorLayoutElementItemGroupFieldType"
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutElementItemGroupFieldType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, EditorInterfaceEditorLayoutElementItemGroupFieldValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutElementItemGroupFieldType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewEditorInterfaceEditorLayoutElementItemGroupFieldValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewEditorInterfaceEditorLayoutElementItemGroupFieldValueNull(), nil
	}

	if !value.IsKnown() {
		return NewEditorInterfaceEditorLayoutElementItemGroupFieldValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create EditorInterfaceEditorLayoutElementItemGroupFieldValue from Terraform: %w", err)
	}

	v, diags := NewEditorInterfaceEditorLayoutElementItemGroupFieldValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutElementItemGroupFieldType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewEditorInterfaceEditorLayoutElementItemGroupFieldValueNull(), nil
	case value.IsUnknown():
		return NewEditorInterfaceEditorLayoutElementItemGroupFieldValueUnknown(), nil
	}

	return NewEditorInterfaceEditorLayoutElementItemGroupFieldValueKnownFromAttributes(ctx, value.Attributes())
}
