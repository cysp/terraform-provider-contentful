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

type ContentTypeFieldAllowedResourceItemExternalValue struct {
	TypeID basetypes.StringValue `tfsdk:"type"`
	state  attr.ValueState
}

func NewContentTypeFieldAllowedResourceItemExternalValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (ContentTypeFieldAllowedResourceItemExternalValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := ContentTypeFieldAllowedResourceItemExternalValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewContentTypeFieldAllowedResourceItemExternalValueNull() ContentTypeFieldAllowedResourceItemExternalValue {
	return ContentTypeFieldAllowedResourceItemExternalValue{
		state: attr.ValueStateNull,
	}
}

func NewContentTypeFieldAllowedResourceItemExternalValueUnknown() ContentTypeFieldAllowedResourceItemExternalValue {
	return ContentTypeFieldAllowedResourceItemExternalValue{
		state: attr.ValueStateUnknown,
	}
}

func (v ContentTypeFieldAllowedResourceItemExternalValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"type": schema.StringAttribute{
			Required: true,
		},
	}
}

//nolint:ireturn
func (v ContentTypeFieldAllowedResourceItemExternalValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return ContentTypeFieldAllowedResourceItemExternalType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = ContentTypeFieldAllowedResourceItemExternalValue{}

//nolint:ireturn
func (v ContentTypeFieldAllowedResourceItemExternalValue) Type(ctx context.Context) attr.Type {
	return ContentTypeFieldAllowedResourceItemExternalType{ObjectType: v.ObjectType(ctx)}
}

func (v ContentTypeFieldAllowedResourceItemExternalValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v ContentTypeFieldAllowedResourceItemExternalValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v ContentTypeFieldAllowedResourceItemExternalValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[ContentTypeFieldAllowedResourceItemExternalValue](v, o)
}

func (v ContentTypeFieldAllowedResourceItemExternalValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ContentTypeFieldAllowedResourceItemExternalValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ContentTypeFieldAllowedResourceItemExternalValue) String() string {
	return "ContentTypeFieldAllowedResourceItemExternalValue"
}

func (v ContentTypeFieldAllowedResourceItemExternalValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v ContentTypeFieldAllowedResourceItemExternalValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
