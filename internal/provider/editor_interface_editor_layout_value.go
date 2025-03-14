package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

//nolint:recvcheck
type EditorInterfaceEditorLayoutValue struct {
	GroupID basetypes.StringValue `tfsdk:"group_id"`
	Name    basetypes.StringValue `tfsdk:"name"`
	Items   basetypes.ListValue   `tfsdk:"items"`
	state   attr.ValueState
}

func NewEditorInterfaceEditorLayoutValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (EditorInterfaceEditorLayoutValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceEditorLayoutValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := setTFSDKAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewEditorInterfaceEditorLayoutValueNull() EditorInterfaceEditorLayoutValue {
	return EditorInterfaceEditorLayoutValue{
		state: attr.ValueStateNull,
	}
}

func NewEditorInterfaceEditorLayoutValueUnknown() EditorInterfaceEditorLayoutValue {
	return EditorInterfaceEditorLayoutValue{
		state: attr.ValueStateUnknown,
	}
}

func (v EditorInterfaceEditorLayoutValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"group_id": schema.StringAttribute{
			Required: true,
		},
		"name": schema.StringAttribute{
			Required: true,
		},
		"items": schema.ListAttribute{
			ElementType: jsontypes.NormalizedType{},
			Optional:    true,
		},
	}
}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return EditorInterfaceEditorLayoutType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = EditorInterfaceEditorLayoutValue{}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutValue) Type(ctx context.Context) attr.Type {
	return EditorInterfaceEditorLayoutType{ObjectType: v.ObjectType(ctx)}
}

func (v EditorInterfaceEditorLayoutValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v EditorInterfaceEditorLayoutValue) ObjectAttrTypes(_ context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"group_id": types.StringType,
		"name":     types.StringType,
		"items":    types.ListType{ElemType: jsontypes.NormalizedType{}},
	}
}

func (v EditorInterfaceEditorLayoutValue) Equal(o attr.Value) bool {
	other, ok := o.(EditorInterfaceEditorLayoutValue)
	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state == attr.ValueStateKnown {
		return compareTFSDKAttributesEqual(v, other)
	}

	return true
}

func (v EditorInterfaceEditorLayoutValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v EditorInterfaceEditorLayoutValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v EditorInterfaceEditorLayoutValue) String() string {
	return "EditorInterfaceEditorLayoutValue"
}

func (v EditorInterfaceEditorLayoutValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	return ReflectToTerraformValue(ctx, v, v.state)
}

func (v EditorInterfaceEditorLayoutValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return ReflectToObjectValue(ctx, v)
}
