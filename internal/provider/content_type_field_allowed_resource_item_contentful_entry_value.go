package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type ContentTypeFieldAllowedResourceItemContentfulEntryValue struct {
	Source       types.String            `tfsdk:"source"`
	ContentTypes TypedList[types.String] `tfsdk:"content_types"`
	state        attr.ValueState
}

func NewContentTypeFieldAllowedResourceItemContentfulEntryValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (ContentTypeFieldAllowedResourceItemContentfulEntryValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := ContentTypeFieldAllowedResourceItemContentfulEntryValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func NewContentTypeFieldAllowedResourceItemContentfulEntryValueNull() ContentTypeFieldAllowedResourceItemContentfulEntryValue {
	return ContentTypeFieldAllowedResourceItemContentfulEntryValue{
		state: attr.ValueStateNull,
	}
}

func NewContentTypeFieldAllowedResourceItemContentfulEntryValueUnknown() ContentTypeFieldAllowedResourceItemContentfulEntryValue {
	return ContentTypeFieldAllowedResourceItemContentfulEntryValue{
		state: attr.ValueStateUnknown,
	}
}

func (v ContentTypeFieldAllowedResourceItemContentfulEntryValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"source": schema.StringAttribute{
			Required: true,
		},
		"content_types": schema.ListAttribute{
			ElementType: types.StringType,
			CustomType:  NewTypedListNull[types.String](ctx).CustomType(ctx),
			Required:    true,
		},
	}
}

//nolint:ireturn
func (v ContentTypeFieldAllowedResourceItemContentfulEntryValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return ContentTypeFieldAllowedResourceItemContentfulEntryType{ObjectType: v.ObjectType(ctx)}
}

var _ basetypes.ObjectValuable = ContentTypeFieldAllowedResourceItemContentfulEntryValue{}

//nolint:ireturn
func (v ContentTypeFieldAllowedResourceItemContentfulEntryValue) Type(ctx context.Context) attr.Type {
	return ContentTypeFieldAllowedResourceItemContentfulEntryType{ObjectType: v.ObjectType(ctx)}
}

func (v ContentTypeFieldAllowedResourceItemContentfulEntryValue) ObjectType(ctx context.Context) basetypes.ObjectType {
	return basetypes.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v ContentTypeFieldAllowedResourceItemContentfulEntryValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v ContentTypeFieldAllowedResourceItemContentfulEntryValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[ContentTypeFieldAllowedResourceItemContentfulEntryValue](v, o)
}

func (v ContentTypeFieldAllowedResourceItemContentfulEntryValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ContentTypeFieldAllowedResourceItemContentfulEntryValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ContentTypeFieldAllowedResourceItemContentfulEntryValue) String() string {
	return "ContentTypeFieldAllowedResourceItemContentfulEntryValue"
}

func (v ContentTypeFieldAllowedResourceItemContentfulEntryValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v ContentTypeFieldAllowedResourceItemContentfulEntryValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
