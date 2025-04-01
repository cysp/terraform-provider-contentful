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
type EditorInterfaceEditorLayoutItemGroupItemValue struct {
	Field EditorInterfaceEditorLayoutItemGroupItemFieldValue `tfsdk:"field"`
	Group EditorInterfaceEditorLayoutItemGroupItemGroupValue `tfsdk:"group"`
	state attr.ValueState
}

func NewEditorInterfaceEditorLayoutItemGroupItemValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (EditorInterfaceEditorLayoutItemGroupItemValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceEditorLayoutItemGroupItemValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewEditorInterfaceEditorLayoutItemGroupItemValueNull() EditorInterfaceEditorLayoutItemGroupItemValue {
	return EditorInterfaceEditorLayoutItemGroupItemValue{
		state: attr.ValueStateNull,
	}
}

func NewEditorInterfaceEditorLayoutItemGroupItemValueUnknown() EditorInterfaceEditorLayoutItemGroupItemValue {
	return EditorInterfaceEditorLayoutItemGroupItemValue{
		state: attr.ValueStateUnknown,
	}
}

func (v EditorInterfaceEditorLayoutItemGroupItemValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"field": schema.SingleNestedAttribute{
			Attributes: EditorInterfaceEditorLayoutItemGroupItemFieldValue{}.SchemaAttributes(ctx),
			CustomType: EditorInterfaceEditorLayoutItemGroupItemFieldValue{}.CustomType(ctx),
			Optional:   true,
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("field"),
					path.MatchRelative().AtParent().AtName("group"),
				),
			},
		},
		"group": schema.SingleNestedAttribute{
			Attributes: EditorInterfaceEditorLayoutItemGroupItemGroupValue{}.SchemaAttributes(ctx),
			CustomType: EditorInterfaceEditorLayoutItemGroupItemGroupValue{}.CustomType(ctx),
			Optional:   true,
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
func (v EditorInterfaceEditorLayoutItemGroupItemValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return EditorInterfaceEditorLayoutItemGroupItemType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = EditorInterfaceEditorLayoutItemGroupItemValue{}

//nolint:ireturn
func (v EditorInterfaceEditorLayoutItemGroupItemValue) Type(ctx context.Context) attr.Type {
	return EditorInterfaceEditorLayoutItemGroupItemType{ObjectType: v.ObjectType(ctx)}
}

func (v EditorInterfaceEditorLayoutItemGroupItemValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v EditorInterfaceEditorLayoutItemGroupItemValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v EditorInterfaceEditorLayoutItemGroupItemValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[EditorInterfaceEditorLayoutItemGroupItemValue](v, o)
}

func (v EditorInterfaceEditorLayoutItemGroupItemValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v EditorInterfaceEditorLayoutItemGroupItemValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v EditorInterfaceEditorLayoutItemGroupItemValue) String() string {
	return "EditorInterfaceEditorLayoutItemGroupItemValue"
}

func (v EditorInterfaceEditorLayoutItemGroupItemValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v EditorInterfaceEditorLayoutItemGroupItemValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
