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
type EditorInterfaceEditorLayoutElementItemGroupValue struct {
	GroupID basetypes.StringValue `tfsdk:"group_id"`
	Name    basetypes.StringValue `tfsdk:"name"`
	Items   basetypes.ListValue   `tfsdk:"items"`
	state   attr.ValueState
}

func NewEditorInterfaceEditorLayoutElementItemGroupValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (EditorInterfaceEditorLayoutElementItemGroupValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceEditorLayoutElementItemGroupValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewEditorInterfaceEditorLayoutElementItemGroupValueNull() EditorInterfaceEditorLayoutElementItemGroupValue {
	return EditorInterfaceEditorLayoutElementItemGroupValue{
		state: attr.ValueStateNull,
	}
}

func NewEditorInterfaceEditorLayoutElementItemGroupValueUnknown() EditorInterfaceEditorLayoutElementItemGroupValue {
	return EditorInterfaceEditorLayoutElementItemGroupValue{
		state: attr.ValueStateUnknown,
	}
}

func (v EditorInterfaceEditorLayoutElementItemGroupValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"group_id": schema.StringAttribute{
			Required: true,
		},
		"name": schema.StringAttribute{
			Required: true,
		},
		"items": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: EditorInterfaceEditorLayoutElementItemGroupFieldValue{}.SchemaAttributes(ctx),
				CustomType: EditorInterfaceEditorLayoutElementItemGroupFieldValue{}.CustomType(ctx),
			},
			Required: true,
		},
	}
}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutElementItemGroupValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return EditorInterfaceEditorLayoutElementItemGroupType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = EditorInterfaceEditorLayoutElementItemGroupValue{}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutElementItemGroupValue) Type(ctx context.Context) attr.Type {
	return EditorInterfaceEditorLayoutElementItemGroupType{ObjectType: v.ObjectType(ctx)}
}

func (v EditorInterfaceEditorLayoutElementItemGroupValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v EditorInterfaceEditorLayoutElementItemGroupValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v EditorInterfaceEditorLayoutElementItemGroupValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[EditorInterfaceEditorLayoutElementItemGroupValue](v, o)
}

func (v EditorInterfaceEditorLayoutElementItemGroupValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v EditorInterfaceEditorLayoutElementItemGroupValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v EditorInterfaceEditorLayoutElementItemGroupValue) String() string {
	return "EditorInterfaceEditorLayoutElementItemGroupValue"
}

func (v EditorInterfaceEditorLayoutElementItemGroupValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v EditorInterfaceEditorLayoutElementItemGroupValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
