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

type EditorInterfaceEditorLayoutItemGroupItemFieldType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = EditorInterfaceEditorLayoutItemGroupItemFieldType{}

func (t EditorInterfaceEditorLayoutItemGroupItemFieldType) Equal(o attr.Type) bool {
	other, ok := o.(EditorInterfaceEditorLayoutItemGroupItemFieldType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t EditorInterfaceEditorLayoutItemGroupItemFieldType) String() string {
	return "EditorInterfaceEditorLayoutItemGroupItemFieldType"
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutItemGroupItemFieldType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, EditorInterfaceEditorLayoutItemGroupItemFieldValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutItemGroupItemFieldType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewEditorInterfaceEditorLayoutItemGroupItemFieldValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewEditorInterfaceEditorLayoutItemGroupItemFieldValueNull(), nil
	}

	if !value.IsKnown() {
		return NewEditorInterfaceEditorLayoutItemGroupItemFieldValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create EditorInterfaceEditorLayoutItemFieldValue from Terraform: %w", err)
	}

	v, diags := NewEditorInterfaceEditorLayoutItemGroupItemFieldKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutItemGroupItemFieldType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewEditorInterfaceEditorLayoutItemGroupItemFieldValueNull(), nil
	case value.IsUnknown():
		return NewEditorInterfaceEditorLayoutItemGroupItemFieldValueUnknown(), nil
	}

	return NewEditorInterfaceEditorLayoutItemGroupItemFieldKnownFromAttributes(ctx, value.Attributes())
}
