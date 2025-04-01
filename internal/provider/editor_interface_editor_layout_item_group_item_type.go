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

type EditorInterfaceEditorLayoutItemGroupItemType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = EditorInterfaceEditorLayoutItemGroupItemType{}

func (t EditorInterfaceEditorLayoutItemGroupItemType) Equal(o attr.Type) bool {
	other, ok := o.(EditorInterfaceEditorLayoutItemGroupItemType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t EditorInterfaceEditorLayoutItemGroupItemType) String() string {
	return "EditorInterfaceEditorLayoutItemGroupItemType"
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutItemGroupItemType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, EditorInterfaceEditorLayoutItemGroupItemValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutItemGroupItemType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewEditorInterfaceEditorLayoutItemGroupItemValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewEditorInterfaceEditorLayoutItemGroupItemValueNull(), nil
	}

	if !value.IsKnown() {
		return NewEditorInterfaceEditorLayoutItemGroupItemValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create EditorInterfaceEditorLayoutItemGroupItemValue from Terraform: %w", err)
	}

	v, diags := NewEditorInterfaceEditorLayoutItemGroupItemValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutItemGroupItemType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewEditorInterfaceEditorLayoutItemGroupItemValueNull(), nil
	case value.IsUnknown():
		return NewEditorInterfaceEditorLayoutItemGroupItemValueUnknown(), nil
	}

	return NewEditorInterfaceEditorLayoutItemGroupItemValueKnownFromAttributes(ctx, value.Attributes())
}
