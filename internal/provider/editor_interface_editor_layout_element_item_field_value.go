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
type EditorInterfaceEditorLayoutElementItemFieldValue struct {
	FieldID basetypes.StringValue `tfsdk:"field_id"`
	state   attr.ValueState
}

func NewEditorInterfaceEditorLayoutElementItemFieldValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (EditorInterfaceEditorLayoutElementItemFieldValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceEditorLayoutElementItemFieldValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewEditorInterfaceEditorLayoutElementItemFieldValueNull() EditorInterfaceEditorLayoutElementItemFieldValue {
	return EditorInterfaceEditorLayoutElementItemFieldValue{
		state: attr.ValueStateNull,
	}
}

func NewEditorInterfaceEditorLayoutElementItemFieldValueUnknown() EditorInterfaceEditorLayoutElementItemFieldValue {
	return EditorInterfaceEditorLayoutElementItemFieldValue{
		state: attr.ValueStateUnknown,
	}
}

func (v EditorInterfaceEditorLayoutElementItemFieldValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"field_id": schema.StringAttribute{
			Required: true,
		},
	}
}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutElementItemFieldValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return EditorInterfaceEditorLayoutElementItemFieldType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = EditorInterfaceEditorLayoutElementItemFieldValue{}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutElementItemFieldValue) Type(ctx context.Context) attr.Type {
	return EditorInterfaceEditorLayoutElementItemFieldType{ObjectType: v.ObjectType(ctx)}
}

func (v EditorInterfaceEditorLayoutElementItemFieldValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v EditorInterfaceEditorLayoutElementItemFieldValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v EditorInterfaceEditorLayoutElementItemFieldValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[EditorInterfaceEditorLayoutElementItemFieldValue](v, o)
}

func (v EditorInterfaceEditorLayoutElementItemFieldValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v EditorInterfaceEditorLayoutElementItemFieldValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v EditorInterfaceEditorLayoutElementItemFieldValue) String() string {
	return "EditorInterfaceEditorLayoutElementItemFieldValue"
}

func (v EditorInterfaceEditorLayoutElementItemFieldValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v EditorInterfaceEditorLayoutElementItemFieldValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
