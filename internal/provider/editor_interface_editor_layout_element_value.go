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
type EditorInterfaceEditorLayoutElementValue struct {
	GroupID basetypes.StringValue `tfsdk:"group_id"`
	Name    basetypes.StringValue `tfsdk:"name"`
	Items   basetypes.ListValue   `tfsdk:"items"`
	state   attr.ValueState
}

func NewEditorInterfaceEditorLayoutElementValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (EditorInterfaceEditorLayoutElementValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceEditorLayoutElementValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewEditorInterfaceEditorLayoutElementValueNull() EditorInterfaceEditorLayoutElementValue {
	return EditorInterfaceEditorLayoutElementValue{
		state: attr.ValueStateNull,
	}
}

func NewEditorInterfaceEditorLayoutElementValueUnknown() EditorInterfaceEditorLayoutElementValue {
	return EditorInterfaceEditorLayoutElementValue{
		state: attr.ValueStateUnknown,
	}
}

func (v EditorInterfaceEditorLayoutElementValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"group_id": schema.StringAttribute{
			Required: true,
		},
		"name": schema.StringAttribute{
			Required: true,
		},
		"items": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: EditorInterfaceEditorLayoutElementItemValue{}.SchemaAttributes(ctx),
				CustomType: EditorInterfaceEditorLayoutElementItemValue{}.CustomType(ctx),
			},
			Required: true,
		},
	}
}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutElementValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return EditorInterfaceEditorLayoutElementType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = EditorInterfaceEditorLayoutElementValue{}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutElementValue) Type(ctx context.Context) attr.Type {
	return EditorInterfaceEditorLayoutElementType{ObjectType: v.ObjectType(ctx)}
}

func (v EditorInterfaceEditorLayoutElementValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v EditorInterfaceEditorLayoutElementValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v EditorInterfaceEditorLayoutElementValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[EditorInterfaceEditorLayoutElementValue](v, o)
}

func (v EditorInterfaceEditorLayoutElementValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v EditorInterfaceEditorLayoutElementValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v EditorInterfaceEditorLayoutElementValue) String() string {
	return "EditorInterfaceEditorLayoutElementValue"
}

func (v EditorInterfaceEditorLayoutElementValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v EditorInterfaceEditorLayoutElementValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
