package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

//nolint:recvcheck
type EditorInterfaceEditorLayoutItemGroupValue struct {
	GroupID types.String                                             `tfsdk:"group_id"`
	Name    types.String                                             `tfsdk:"name"`
	Items   TypedList[EditorInterfaceEditorLayoutItemGroupItemValue] `tfsdk:"items"`
	state   attr.ValueState
}

func NewEditorInterfaceEditorLayoutItemGroupValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (EditorInterfaceEditorLayoutItemGroupValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceEditorLayoutItemGroupValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewEditorInterfaceEditorLayoutItemGroupValueNull() EditorInterfaceEditorLayoutItemGroupValue {
	return EditorInterfaceEditorLayoutItemGroupValue{
		state: attr.ValueStateNull,
	}
}

func NewEditorInterfaceEditorLayoutItemGroupValueUnknown() EditorInterfaceEditorLayoutItemGroupValue {
	return EditorInterfaceEditorLayoutItemGroupValue{
		state: attr.ValueStateUnknown,
	}
}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutItemGroupValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return EditorInterfaceEditorLayoutItemGroupType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = EditorInterfaceEditorLayoutItemGroupValue{}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutItemGroupValue) Type(ctx context.Context) attr.Type {
	return EditorInterfaceEditorLayoutItemGroupType{ObjectType: v.ObjectType(ctx)}
}

func (v EditorInterfaceEditorLayoutItemGroupValue) ObjectType(ctx context.Context) types.ObjectType {
	return types.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v EditorInterfaceEditorLayoutItemGroupValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v EditorInterfaceEditorLayoutItemGroupValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[EditorInterfaceEditorLayoutItemGroupValue](v, o)
}

func (v EditorInterfaceEditorLayoutItemGroupValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v EditorInterfaceEditorLayoutItemGroupValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v EditorInterfaceEditorLayoutItemGroupValue) String() string {
	return "EditorInterfaceEditorLayoutItemGroupValue"
}

func (v EditorInterfaceEditorLayoutItemGroupValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v EditorInterfaceEditorLayoutItemGroupValue) ToObjectValue(ctx context.Context) (types.Object, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
