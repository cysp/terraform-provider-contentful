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
type EditorInterfaceGroupControlValue struct {
	GroupID         types.String         `tfsdk:"group_id"`
	WidgetNamespace types.String         `tfsdk:"widget_namespace"`
	WidgetID        types.String         `tfsdk:"widget_id"`
	Settings        jsontypes.Normalized `tfsdk:"settings"`
	state           attr.ValueState
}

func NewEditorInterfaceGroupControlValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (EditorInterfaceGroupControlValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceGroupControlValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewEditorInterfaceGroupControlValueNull() EditorInterfaceGroupControlValue {
	return EditorInterfaceGroupControlValue{
		state: attr.ValueStateNull,
	}
}

func NewEditorInterfaceGroupControlValueUnknown() EditorInterfaceGroupControlValue {
	return EditorInterfaceGroupControlValue{
		state: attr.ValueStateUnknown,
	}
}

//nolint:ireturn
func (v EditorInterfaceGroupControlValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return EditorInterfaceGroupControlType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = EditorInterfaceGroupControlValue{}

//nolint:ireturn
func (v EditorInterfaceGroupControlValue) Type(ctx context.Context) attr.Type {
	return EditorInterfaceGroupControlType{ObjectType: v.ObjectType(ctx)}
}

func (v EditorInterfaceGroupControlValue) ObjectType(ctx context.Context) types.ObjectType {
	return types.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v EditorInterfaceGroupControlValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v EditorInterfaceGroupControlValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[EditorInterfaceGroupControlValue](v, o)
}

func (v EditorInterfaceGroupControlValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v EditorInterfaceGroupControlValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v EditorInterfaceGroupControlValue) String() string {
	return "EditorInterfaceGroupControlValue"
}

func (v EditorInterfaceGroupControlValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v EditorInterfaceGroupControlValue) ToObjectValue(ctx context.Context) (types.Object, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
