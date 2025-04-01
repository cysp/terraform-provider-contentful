package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

//nolint:recvcheck
type EditorInterfaceEditorLayoutItemGroupItemGroupItemValue struct {
	Field EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue `tfsdk:"field"`
	state attr.ValueState
}

func NewEditorInterfaceEditorLayoutItemGroupItemGroupItemValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (EditorInterfaceEditorLayoutItemGroupItemGroupItemValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceEditorLayoutItemGroupItemGroupItemValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewEditorInterfaceEditorLayoutItemGroupItemGroupItemValueNull() EditorInterfaceEditorLayoutItemGroupItemGroupItemValue {
	return EditorInterfaceEditorLayoutItemGroupItemGroupItemValue{
		state: attr.ValueStateNull,
	}
}

func NewEditorInterfaceEditorLayoutItemGroupItemGroupItemValueUnknown() EditorInterfaceEditorLayoutItemGroupItemGroupItemValue {
	return EditorInterfaceEditorLayoutItemGroupItemGroupItemValue{
		state: attr.ValueStateUnknown,
	}
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"field": schema.SingleNestedAttribute{
			Attributes: EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue{}.SchemaAttributes(ctx),
			CustomType: EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue{}.CustomType(ctx),
			Required:   true,
		},
	}
}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return EditorInterfaceEditorLayoutItemGroupItemGroupItemType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = EditorInterfaceEditorLayoutItemGroupItemGroupItemValue{}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemValue) Type(ctx context.Context) attr.Type {
	return EditorInterfaceEditorLayoutItemGroupItemGroupItemType{ObjectType: v.ObjectType(ctx)}
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue](v, o)
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemValue) String() string {
	return "EditorInterfaceEditorLayoutItemGroupItemGroupItemValue"
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
