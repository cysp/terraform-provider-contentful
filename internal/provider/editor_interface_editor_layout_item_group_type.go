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

type EditorInterfaceEditorLayoutItemGroupType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = EditorInterfaceEditorLayoutItemGroupType{}

func (t EditorInterfaceEditorLayoutItemGroupType) Equal(o attr.Type) bool {
	other, ok := o.(EditorInterfaceEditorLayoutItemGroupType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t EditorInterfaceEditorLayoutItemGroupType) String() string {
	return "EditorInterfaceEditorLayoutItemGroupType"
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutItemGroupType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, EditorInterfaceEditorLayoutItemGroupValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutItemGroupType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewEditorInterfaceEditorLayoutItemGroupValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewEditorInterfaceEditorLayoutItemGroupValueNull(), nil
	}

	if !value.IsKnown() {
		return NewEditorInterfaceEditorLayoutItemGroupValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create EditorInterfaceEditorLayoutItemGroupValue from Terraform: %w", err)
	}

	v, diags := NewEditorInterfaceEditorLayoutItemGroupValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t EditorInterfaceEditorLayoutItemGroupType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewEditorInterfaceEditorLayoutItemGroupValueNull(), nil
	case value.IsUnknown():
		return NewEditorInterfaceEditorLayoutItemGroupValueUnknown(), nil
	}

	return NewEditorInterfaceEditorLayoutItemGroupValueKnownFromAttributes(ctx, value.Attributes())
}
