package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

//nolint:recvcheck
type EditorInterfaceSidebarValue struct {
	WidgetNamespace types.String         `tfsdk:"widget_namespace"`
	WidgetID        types.String         `tfsdk:"widget_id"`
	Settings        jsontypes.Normalized `tfsdk:"settings"`
	Disabled        types.Bool           `tfsdk:"disabled"` // updated from field_id to disabled
	state           attr.ValueState
}

func NewEditorInterfaceSidebarValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (EditorInterfaceSidebarValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := EditorInterfaceSidebarValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewEditorInterfaceSidebarValueNull() EditorInterfaceSidebarValue {
	return EditorInterfaceSidebarValue{
		state: attr.ValueStateNull,
	}
}

func NewEditorInterfaceSidebarValueUnknown() EditorInterfaceSidebarValue {
	return EditorInterfaceSidebarValue{
		state: attr.ValueStateUnknown,
	}
}

func (v EditorInterfaceSidebarValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"widget_namespace": schema.StringAttribute{
			Required: true,
		},
		"widget_id": schema.StringAttribute{
			Required: true,
		},
		"settings": schema.StringAttribute{
			CustomType: jsontypes.NormalizedType{},
			Optional:   true,
		},
		"disabled": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
		},
	}
}

//nolint:ireturn
func (v EditorInterfaceSidebarValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return EditorInterfaceSidebarType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = EditorInterfaceSidebarValue{}

//nolint:ireturn
func (v EditorInterfaceSidebarValue) Type(ctx context.Context) attr.Type {
	return EditorInterfaceSidebarType{ObjectType: v.ObjectType(ctx)}
}

func (v EditorInterfaceSidebarValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v EditorInterfaceSidebarValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v EditorInterfaceSidebarValue) Equal(o attr.Value) bool {
	other, ok := o.(EditorInterfaceSidebarValue)
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

func (v EditorInterfaceSidebarValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v EditorInterfaceSidebarValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v EditorInterfaceSidebarValue) String() string {
	return "EditorInterfaceSidebarValue"
}

func (v EditorInterfaceSidebarValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	return ReflectToTerraformValue(ctx, v, v.state)
}

func (v EditorInterfaceSidebarValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return ReflectToObjectValue(ctx, v)
}
