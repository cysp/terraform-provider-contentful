package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

//nolint:recvcheck
type EditorInterfaceControlValue struct {
	FieldID         types.String         `tfsdk:"field_id"`
	WidgetNamespace types.String         `tfsdk:"widget_namespace"`
	WidgetID        types.String         `tfsdk:"widget_id"`
	Settings        jsontypes.Normalized `tfsdk:"settings"`
	state           attr.ValueState
}

func NewEditorInterfaceControlValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (EditorInterfaceControlValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceControlValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewEditorInterfaceControlValueNull() EditorInterfaceControlValue {
	return EditorInterfaceControlValue{
		state: attr.ValueStateNull,
	}
}

func NewEditorInterfaceControlValueUnknown() EditorInterfaceControlValue {
	return EditorInterfaceControlValue{
		state: attr.ValueStateUnknown,
	}
}

//nolint:ireturn
func (v EditorInterfaceControlValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return EditorInterfaceControlType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = EditorInterfaceControlValue{}

//nolint:ireturn
func (v EditorInterfaceControlValue) Type(ctx context.Context) attr.Type {
	return EditorInterfaceControlType{ObjectType: v.ObjectType(ctx)}
}

func (v EditorInterfaceControlValue) ObjectType(ctx context.Context) types.ObjectType {
	return types.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v EditorInterfaceControlValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v EditorInterfaceControlValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[EditorInterfaceControlValue](v, o)
}

func (v EditorInterfaceControlValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v EditorInterfaceControlValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v EditorInterfaceControlValue) String() string {
	return "EditorInterfaceControlValue"
}

func (v EditorInterfaceControlValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v EditorInterfaceControlValue) ToObjectValue(ctx context.Context) (types.Object, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
