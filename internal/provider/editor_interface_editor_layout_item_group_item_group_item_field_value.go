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
type EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue struct {
	FieldID basetypes.StringValue `tfsdk:"field_id"`
	state   attr.ValueState
}

func NewEditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewEditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValueNull() EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue {
	return EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue{
		state: attr.ValueStateNull,
	}
}

func NewEditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValueUnknown() EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue {
	return EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue{
		state: attr.ValueStateUnknown,
	}
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"field_id": schema.StringAttribute{
			Required: true,
		},
	}
}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue{}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue) Type(ctx context.Context) attr.Type {
	return EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldType{ObjectType: v.ObjectType(ctx)}
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue](v, o)
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue) String() string {
	return "EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue"
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
