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
type EditorInterfaceEditorLayoutElementItemGroupFieldValue struct {
	FieldID basetypes.StringValue `tfsdk:"field_id"`
	state   attr.ValueState
}

func NewEditorInterfaceEditorLayoutElementItemGroupFieldValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (EditorInterfaceEditorLayoutElementItemGroupFieldValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceEditorLayoutElementItemGroupFieldValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewEditorInterfaceEditorLayoutElementItemGroupFieldValueNull() EditorInterfaceEditorLayoutElementItemGroupFieldValue {
	return EditorInterfaceEditorLayoutElementItemGroupFieldValue{
		state: attr.ValueStateNull,
	}
}

func NewEditorInterfaceEditorLayoutElementItemGroupFieldValueUnknown() EditorInterfaceEditorLayoutElementItemGroupFieldValue {
	return EditorInterfaceEditorLayoutElementItemGroupFieldValue{
		state: attr.ValueStateUnknown,
	}
}

func (v EditorInterfaceEditorLayoutElementItemGroupFieldValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"field_id": schema.StringAttribute{
			Required: true,
		},
	}
}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutElementItemGroupFieldValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return EditorInterfaceEditorLayoutElementItemGroupFieldType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = EditorInterfaceEditorLayoutElementItemGroupFieldValue{}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutElementItemGroupFieldValue) Type(ctx context.Context) attr.Type {
	return EditorInterfaceEditorLayoutElementItemGroupFieldType{ObjectType: v.ObjectType(ctx)}
}

func (v EditorInterfaceEditorLayoutElementItemGroupFieldValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v EditorInterfaceEditorLayoutElementItemGroupFieldValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v EditorInterfaceEditorLayoutElementItemGroupFieldValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[EditorInterfaceEditorLayoutElementItemGroupFieldValue](v, o)
}

func (v EditorInterfaceEditorLayoutElementItemGroupFieldValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v EditorInterfaceEditorLayoutElementItemGroupFieldValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v EditorInterfaceEditorLayoutElementItemGroupFieldValue) String() string {
	return "EditorInterfaceEditorLayoutElementItemGroupFieldValue"
}

func (v EditorInterfaceEditorLayoutElementItemGroupFieldValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v EditorInterfaceEditorLayoutElementItemGroupFieldValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
