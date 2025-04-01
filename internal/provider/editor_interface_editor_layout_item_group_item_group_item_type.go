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

type EditorInterfaceEditorLayoutItemGroupItemGroupItemType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = EditorInterfaceEditorLayoutItemGroupItemGroupItemType{}

func (t EditorInterfaceEditorLayoutItemGroupItemGroupItemType) Equal(o attr.Type) bool {
	other, ok := o.(EditorInterfaceEditorLayoutItemGroupItemGroupItemType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t EditorInterfaceEditorLayoutItemGroupItemGroupItemType) String() string {
	return "EditorInterfaceEditorLayoutItemGroupItemGroupItemType"
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutItemGroupItemGroupItemType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, EditorInterfaceEditorLayoutItemGroupItemGroupItemValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutItemGroupItemGroupItemType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewEditorInterfaceEditorLayoutItemGroupItemGroupItemValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewEditorInterfaceEditorLayoutItemGroupItemGroupItemValueNull(), nil
	}

	if !value.IsKnown() {
		return NewEditorInterfaceEditorLayoutItemGroupItemGroupItemValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create EditorInterfaceEditorLayoutItemGroupItemGroupItemValue from Terraform: %w", err)
	}

	v, diags := NewEditorInterfaceEditorLayoutItemGroupItemGroupItemValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutItemGroupItemGroupItemType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewEditorInterfaceEditorLayoutItemGroupItemGroupItemValueNull(), nil
	case value.IsUnknown():
		return NewEditorInterfaceEditorLayoutItemGroupItemGroupItemValueUnknown(), nil
	}

	return NewEditorInterfaceEditorLayoutItemGroupItemGroupItemValueKnownFromAttributes(ctx, value.Attributes())
}
