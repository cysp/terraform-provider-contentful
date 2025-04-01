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
type EditorInterfaceEditorLayoutItemGroupItemGroupValue struct {
	GroupID basetypes.StringValue                                             `tfsdk:"group_id"`
	Name    basetypes.StringValue                                             `tfsdk:"name"`
	Items   TypedList[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue] `tfsdk:"items"`
	state   attr.ValueState
}

func NewEditorInterfaceEditorLayoutItemGroupItemGroupValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (EditorInterfaceEditorLayoutItemGroupItemGroupValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceEditorLayoutItemGroupItemGroupValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewEditorInterfaceEditorLayoutItemGroupItemGroupValueNull() EditorInterfaceEditorLayoutItemGroupItemGroupValue {
	return EditorInterfaceEditorLayoutItemGroupItemGroupValue{
		state: attr.ValueStateNull,
	}
}

func NewEditorInterfaceEditorLayoutItemGroupItemGroupValueUnknown() EditorInterfaceEditorLayoutItemGroupItemGroupValue {
	return EditorInterfaceEditorLayoutItemGroupItemGroupValue{
		state: attr.ValueStateUnknown,
	}
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"group_id": schema.StringAttribute{
			Required: true,
		},
		"name": schema.StringAttribute{
			Required: true,
		},
		"items": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: EditorInterfaceEditorLayoutItemGroupItemGroupItemValue{}.SchemaAttributes(ctx),
				CustomType: EditorInterfaceEditorLayoutItemGroupItemGroupItemValue{}.CustomType(ctx),
			},
			CustomType: NewTypedListNull[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue](ctx).CustomType(ctx),
			Required:   true,
		},
	}
}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutItemGroupItemGroupValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return EditorInterfaceEditorLayoutItemGroupItemGroupType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = EditorInterfaceEditorLayoutItemGroupItemGroupValue{}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutItemGroupItemGroupValue) Type(ctx context.Context) attr.Type {
	return EditorInterfaceEditorLayoutItemGroupItemGroupType{ObjectType: v.ObjectType(ctx)}
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[EditorInterfaceEditorLayoutItemGroupItemGroupValue](v, o)
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupValue) String() string {
	return "EditorInterfaceEditorLayoutItemGroupItemGroupValue"
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
