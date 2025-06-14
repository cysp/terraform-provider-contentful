package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

//nolint:recvcheck
type EditorInterfaceEditorLayoutItemValue struct {
	Group EditorInterfaceEditorLayoutItemGroupValue `tfsdk:"group"`
	state attr.ValueState
}

func NewEditorInterfaceEditorLayoutItemValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (EditorInterfaceEditorLayoutItemValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceEditorLayoutItemValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewEditorInterfaceEditorLayoutItemValueNull() EditorInterfaceEditorLayoutItemValue {
	return EditorInterfaceEditorLayoutItemValue{
		state: attr.ValueStateNull,
	}
}

func NewEditorInterfaceEditorLayoutItemValueUnknown() EditorInterfaceEditorLayoutItemValue {
	return EditorInterfaceEditorLayoutItemValue{
		state: attr.ValueStateUnknown,
	}
}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutItemValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return EditorInterfaceEditorLayoutItemType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = EditorInterfaceEditorLayoutItemValue{}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutItemValue) Type(ctx context.Context) attr.Type {
	return EditorInterfaceEditorLayoutItemType{ObjectType: v.ObjectType(ctx)}
}

func (v EditorInterfaceEditorLayoutItemValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v EditorInterfaceEditorLayoutItemValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v EditorInterfaceEditorLayoutItemValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[EditorInterfaceEditorLayoutItemValue](v, o)
}

func (v EditorInterfaceEditorLayoutItemValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v EditorInterfaceEditorLayoutItemValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v EditorInterfaceEditorLayoutItemValue) String() string {
	return "EditorInterfaceEditorLayoutItemValue"
}

func (v EditorInterfaceEditorLayoutItemValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v EditorInterfaceEditorLayoutItemValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
