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
type EditorInterfaceControlValue struct {
	FieldID         types.String         `tfsdk:"field_id"`
	WidgetNamespace types.String         `tfsdk:"widget_namespace"`
	WidgetID        types.String         `tfsdk:"widget_id"`
	Settings        jsontypes.Normalized `tfsdk:"settings"`
	state           attr.ValueState
}

func NewEditorInterfaceControlValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (EditorInterfaceControlValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceControlValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := setTFSDKAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewEditorInterfaceControlValueNull() EditorInterfaceControlValue {
	return EditorInterfaceControlValue{
		state: attr.ValueStateNull,
	}
}

func NewEditorInterfaceControlValueUnknown() EditorInterfaceControlValue {
	return EditorInterfaceControlValue{
		state: attr.ValueStateUnknown,
	}
}

func (v EditorInterfaceControlValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"field_id": schema.StringAttribute{
			Required: true,
		},
		"widget_namespace": schema.StringAttribute{
			Optional: true,
		},
		"widget_id": schema.StringAttribute{
			Optional: true,
		},
		"settings": schema.StringAttribute{
			CustomType: jsontypes.NormalizedType{},
			Optional:   true,
		},
	}
}

//nolint:ireturn
func (v EditorInterfaceControlValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return EditorInterfaceControlType{
		v.ObjectType(ctx),
	}
}

var _ basetypes.ObjectValuable = EditorInterfaceControlValue{}

//nolint:ireturn
func (v EditorInterfaceControlValue) Type(ctx context.Context) attr.Type {
	return EditorInterfaceControlType{
		ObjectType: v.ObjectType(ctx),
	}
}

func (v EditorInterfaceControlValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{
		AttrTypes: v.ObjectAttrTypes(ctx),
	}
}

func (v EditorInterfaceControlValue) ObjectAttrTypes(_ context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"field_id":         types.StringType,
		"widget_namespace": types.StringType,
		"widget_id":        types.StringType,
		"settings":         jsontypes.NormalizedType{},
	}
}

func (v EditorInterfaceControlValue) Equal(o attr.Value) bool {
	other, ok := o.(EditorInterfaceControlValue)
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

func (v EditorInterfaceControlValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v EditorInterfaceControlValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v EditorInterfaceControlValue) String() string {
	return "EditorInterfaceControlValue"
}

func (v EditorInterfaceControlValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	return ReflectToTerraformValue(ctx, v, v.state)
}

func (v EditorInterfaceControlValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return ReflectToObjectValue(ctx, v)
}
