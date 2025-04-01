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

type EditorInterfaceEditorLayoutItemGroupItemGroupType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = EditorInterfaceEditorLayoutItemGroupItemGroupType{}

func (t EditorInterfaceEditorLayoutItemGroupItemGroupType) Equal(o attr.Type) bool {
	other, ok := o.(EditorInterfaceEditorLayoutItemGroupItemGroupType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t EditorInterfaceEditorLayoutItemGroupItemGroupType) String() string {
	return "EditorInterfaceEditorLayoutItemGroupItemGroupType"
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutItemGroupItemGroupType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, EditorInterfaceEditorLayoutItemGroupItemGroupValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutItemGroupItemGroupType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewEditorInterfaceEditorLayoutItemGroupItemGroupValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewEditorInterfaceEditorLayoutItemGroupItemGroupValueNull(), nil
	}

	if !value.IsKnown() {
		return NewEditorInterfaceEditorLayoutItemGroupItemGroupValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create EditorInterfaceEditorLayoutItemGroupItemGroupValue from Terraform: %w", err)
	}

	v, diags := NewEditorInterfaceEditorLayoutItemGroupItemGroupValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutItemGroupItemGroupType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewEditorInterfaceEditorLayoutItemGroupItemGroupValueNull(), nil
	case value.IsUnknown():
		return NewEditorInterfaceEditorLayoutItemGroupItemGroupValueUnknown(), nil
	}

	return NewEditorInterfaceEditorLayoutItemGroupItemGroupValueKnownFromAttributes(ctx, value.Attributes())
}
