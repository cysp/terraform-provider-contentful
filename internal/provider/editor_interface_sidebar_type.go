package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type EditorInterfaceSidebarType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = EditorInterfaceSidebarType{}

func (t EditorInterfaceSidebarType) Equal(o attr.Type) bool {
	other, ok := o.(EditorInterfaceSidebarType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

//nolint:ireturn
func (t EditorInterfaceSidebarType) ValueType(_ context.Context) attr.Value {
	return EditorInterfaceSidebarValue{}
}

func (t EditorInterfaceSidebarType) String() string {
	return "EditorInterfaceSidebarType"
}

//nolint:ireturn
func (t EditorInterfaceSidebarType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: ObjectAttrTypesToTerraformTypes(ctx, EditorInterfaceSidebarValue{}.ObjectAttrTypes(ctx)),
	}
}

//nolint:ireturn
func (t EditorInterfaceSidebarType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewEditorInterfaceSidebarValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewEditorInterfaceSidebarValueNull(), nil
	}

	if !value.IsKnown() {
		return NewEditorInterfaceSidebarValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create EditorInterfaceSidebarValue from Terraform: %w", err)
	}

	v, diags := NewEditorInterfaceSidebarValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t EditorInterfaceSidebarType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewEditorInterfaceSidebarValueNull(), nil
	case value.IsUnknown():
		return NewEditorInterfaceSidebarValueUnknown(), nil
	}

	return NewEditorInterfaceSidebarValueKnownFromAttributes(ctx, value.Attributes())
}
