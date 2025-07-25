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
type EditorInterfaceSidebarValue struct {
	WidgetNamespace types.String         `tfsdk:"widget_namespace"`
	WidgetID        types.String         `tfsdk:"widget_id"`
	Settings        jsontypes.Normalized `tfsdk:"settings"`
	Disabled        types.Bool           `tfsdk:"disabled"` // updated from field_id to disabled
	state           attr.ValueState
}

func NewEditorInterfaceSidebarValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (EditorInterfaceSidebarValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceSidebarValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewEditorInterfaceSidebarValueNull() EditorInterfaceSidebarValue {
	return EditorInterfaceSidebarValue{
		state: attr.ValueStateNull,
	}
}

func NewEditorInterfaceSidebarValueUnknown() EditorInterfaceSidebarValue {
	return EditorInterfaceSidebarValue{
		state: attr.ValueStateUnknown,
	}
}

//nolint:ireturn
func (v EditorInterfaceSidebarValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return EditorInterfaceSidebarType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = EditorInterfaceSidebarValue{}

//nolint:ireturn
func (v EditorInterfaceSidebarValue) Type(ctx context.Context) attr.Type {
	return EditorInterfaceSidebarType{ObjectType: v.ObjectType(ctx)}
}

func (v EditorInterfaceSidebarValue) ObjectType(ctx context.Context) types.ObjectType {
	return types.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v EditorInterfaceSidebarValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v EditorInterfaceSidebarValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[EditorInterfaceSidebarValue](v, o)
}

func (v EditorInterfaceSidebarValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v EditorInterfaceSidebarValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v EditorInterfaceSidebarValue) String() string {
	return "EditorInterfaceSidebarValue"
}

func (v EditorInterfaceSidebarValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v EditorInterfaceSidebarValue) ToObjectValue(ctx context.Context) (types.Object, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
