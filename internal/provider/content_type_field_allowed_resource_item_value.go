package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

//nolint:recvcheck
type ContentTypeFieldAllowedResourceItemValue struct {
	ContentfulEntry ContentTypeFieldAllowedResourceItemContentfulEntryValue `tfsdk:"contentful_entry"`
	External        ContentTypeFieldAllowedResourceItemExternalValue        `tfsdk:"external"`
	state           attr.ValueState
}

func NewContentTypeFieldAllowedResourceItemValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (ContentTypeFieldAllowedResourceItemValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := ContentTypeFieldAllowedResourceItemValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewContentTypeFieldAllowedResourceItemValueNull() ContentTypeFieldAllowedResourceItemValue {
	return ContentTypeFieldAllowedResourceItemValue{
		state: attr.ValueStateNull,
	}
}

func NewContentTypeFieldAllowedResourceItemValueUnknown() ContentTypeFieldAllowedResourceItemValue {
	return ContentTypeFieldAllowedResourceItemValue{
		state: attr.ValueStateUnknown,
	}
}

//nolint:ireturn
func (v ContentTypeFieldAllowedResourceItemValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return ContentTypeFieldAllowedResourceItemType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = ContentTypeFieldAllowedResourceItemValue{}

//nolint:ireturn
func (v ContentTypeFieldAllowedResourceItemValue) Type(ctx context.Context) attr.Type {
	return ContentTypeFieldAllowedResourceItemType{ObjectType: v.ObjectType(ctx)}
}

func (v ContentTypeFieldAllowedResourceItemValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v ContentTypeFieldAllowedResourceItemValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v ContentTypeFieldAllowedResourceItemValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[ContentTypeFieldAllowedResourceItemValue](v, o)
}

func (v ContentTypeFieldAllowedResourceItemValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ContentTypeFieldAllowedResourceItemValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ContentTypeFieldAllowedResourceItemValue) String() string {
	return "ContentTypeFieldAllowedResourceItemValue"
}

func (v ContentTypeFieldAllowedResourceItemValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v ContentTypeFieldAllowedResourceItemValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
