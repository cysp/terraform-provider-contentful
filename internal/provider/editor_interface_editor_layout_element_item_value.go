package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

//nolint:recvcheck
type EditorInterfaceEditorLayoutElementItemValue struct {
	Field EditorInterfaceEditorLayoutElementItemFieldValue `tfsdk:"field"`
	Group EditorInterfaceEditorLayoutElementItemGroupValue `tfsdk:"group"`
	state attr.ValueState
}

func NewEditorInterfaceEditorLayoutElementItemValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (EditorInterfaceEditorLayoutElementItemValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceEditorLayoutElementItemValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewEditorInterfaceEditorLayoutElementItemValueNull() EditorInterfaceEditorLayoutElementItemValue {
	return EditorInterfaceEditorLayoutElementItemValue{
		state: attr.ValueStateNull,
	}
}

func NewEditorInterfaceEditorLayoutElementItemValueUnknown() EditorInterfaceEditorLayoutElementItemValue {
	return EditorInterfaceEditorLayoutElementItemValue{
		state: attr.ValueStateUnknown,
	}
}

func (v EditorInterfaceEditorLayoutElementItemValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"field": schema.ObjectAttribute{
			AttributeTypes: EditorInterfaceEditorLayoutElementItemFieldValue{}.ObjectAttrTypes(ctx),
			CustomType:     EditorInterfaceEditorLayoutElementItemFieldValue{}.CustomType(ctx),
			Optional:       true,
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("field"),
					path.MatchRelative().AtParent().AtName("group"),
				),
			},
		},
		"group": schema.ObjectAttribute{
			AttributeTypes: EditorInterfaceEditorLayoutElementItemGroupValue{}.ObjectAttrTypes(ctx),
			CustomType:     EditorInterfaceEditorLayoutElementItemGroupValue{}.CustomType(ctx),
			Optional:       true,
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("field"),
					path.MatchRelative().AtParent().AtName("group"),
				),
			},
		},
	}
}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutElementItemValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return EditorInterfaceEditorLayoutElementItemType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = EditorInterfaceEditorLayoutElementItemValue{}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutElementItemValue) Type(ctx context.Context) attr.Type {
	return EditorInterfaceEditorLayoutElementItemType{ObjectType: v.ObjectType(ctx)}
}

func (v EditorInterfaceEditorLayoutElementItemValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v EditorInterfaceEditorLayoutElementItemValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v EditorInterfaceEditorLayoutElementItemValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[EditorInterfaceEditorLayoutElementItemValue](v, o)
}

func (v EditorInterfaceEditorLayoutElementItemValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v EditorInterfaceEditorLayoutElementItemValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v EditorInterfaceEditorLayoutElementItemValue) String() string {
	return "EditorInterfaceEditorLayoutElementItemValue"
}

func (v EditorInterfaceEditorLayoutElementItemValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v EditorInterfaceEditorLayoutElementItemValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
