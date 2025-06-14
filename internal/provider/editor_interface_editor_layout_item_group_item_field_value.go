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
type EditorInterfaceEditorLayoutItemGroupItemFieldValue struct {
	FieldID basetypes.StringValue `tfsdk:"field_id"`
	state   attr.ValueState
}

func NewEditorInterfaceEditorLayoutItemGroupItemFieldKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (EditorInterfaceEditorLayoutItemGroupItemFieldValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceEditorLayoutItemGroupItemFieldValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewEditorInterfaceEditorLayoutItemGroupItemFieldValueNull() EditorInterfaceEditorLayoutItemGroupItemFieldValue {
	return EditorInterfaceEditorLayoutItemGroupItemFieldValue{
		state: attr.ValueStateNull,
	}
}

func NewEditorInterfaceEditorLayoutItemGroupItemFieldValueUnknown() EditorInterfaceEditorLayoutItemGroupItemFieldValue {
	return EditorInterfaceEditorLayoutItemGroupItemFieldValue{
		state: attr.ValueStateUnknown,
	}
}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutItemGroupItemFieldValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return EditorInterfaceEditorLayoutItemGroupItemFieldType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = EditorInterfaceEditorLayoutItemGroupItemFieldValue{}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutItemGroupItemFieldValue) Type(ctx context.Context) attr.Type {
	return EditorInterfaceEditorLayoutItemGroupItemFieldType{ObjectType: v.ObjectType(ctx)}
}

func (v EditorInterfaceEditorLayoutItemGroupItemFieldValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v EditorInterfaceEditorLayoutItemGroupItemFieldValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v EditorInterfaceEditorLayoutItemGroupItemFieldValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[EditorInterfaceEditorLayoutItemGroupItemFieldValue](v, o)
}

func (v EditorInterfaceEditorLayoutItemGroupItemFieldValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v EditorInterfaceEditorLayoutItemGroupItemFieldValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v EditorInterfaceEditorLayoutItemGroupItemFieldValue) String() string {
	return "EditorInterfaceEditorLayoutItemFieldValue"
}

func (v EditorInterfaceEditorLayoutItemGroupItemFieldValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v EditorInterfaceEditorLayoutItemGroupItemFieldValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
