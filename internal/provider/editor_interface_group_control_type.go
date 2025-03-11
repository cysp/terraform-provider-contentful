package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type EditorInterfaceGroupControlType struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = EditorInterfaceGroupControlType{}

func (t EditorInterfaceGroupControlType) Equal(o attr.Type) bool {
	other, ok := o.(EditorInterfaceGroupControlType)
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

//nolint:ireturn
func (t EditorInterfaceGroupControlType) ValueType(_ context.Context) attr.Value {
	return EditorInterfaceGroupControlValue{}
}

func (t EditorInterfaceGroupControlType) String() string {
	return "EditorInterfaceGroupControlType"
}

//nolint:ireturn
func (t EditorInterfaceGroupControlType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.Object{
		AttributeTypes: t.TerraformAttributeTypes(ctx),
	}
}

func (t EditorInterfaceGroupControlType) TerraformAttributeTypes(ctx context.Context) map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"group_id":         types.StringType.TerraformType(ctx),
		"widget_namespace": types.StringType.TerraformType(ctx),
		"widget_id":        types.StringType.TerraformType(ctx),
		"settings":         jsontypes.NormalizedType{}.TerraformType(ctx),
	}
}

//nolint:ireturn
func (t EditorInterfaceGroupControlType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.Type() == nil {
		return NewEditorInterfaceGroupControlValueNull(), nil
	}

	if !value.Type().Equal(t.TerraformType(ctx)) {
		return nil, UnexpectedTerraformTypeError{Expected: t.TerraformType(ctx), Actual: value.Type()}
	}

	if value.IsNull() {
		return NewEditorInterfaceGroupControlValueNull(), nil
	}

	if !value.IsKnown() {
		return NewEditorInterfaceGroupControlValueUnknown(), nil
	}

	attributes, err := AttributesFromTerraformValue(ctx, t.AttrTypes, value)
	if err != nil {
		return nil, fmt.Errorf("failed to create EditorInterfaceGroupControlValue from Terraform: %w", err)
	}

	v, diags := NewEditorInterfaceGroupControlValueKnownFromAttributes(ctx, attributes)

	return v, ErrorFromDiags(diags)
}

//nolint:ireturn
func (t EditorInterfaceGroupControlType) ValueFromObject(ctx context.Context, value basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	switch {
	case value.IsNull():
		return NewEditorInterfaceGroupControlValueNull(), nil
	case value.IsUnknown():
		return NewEditorInterfaceGroupControlValueUnknown(), nil
	}

	return NewEditorInterfaceGroupControlValueKnownFromAttributes(ctx, value.Attributes())
}
