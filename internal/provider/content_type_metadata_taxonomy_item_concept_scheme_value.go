package provider

import (
	"context"

	tpfr "github.com/cysp/terraform-provider-contentful/internal/terraform-plugin-framework-reflection"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type ContentTypeMetadataTaxonomyItemConceptSchemeValue struct {
	ID       types.String `tfsdk:"id"`
	Required types.Bool   `tfsdk:"required"`
	state    attr.ValueState
}

var _ basetypes.ObjectValuable = ContentTypeMetadataTaxonomyItemConceptSchemeValue{}

func NewContentTypeMetadataTaxonomyItemConceptSchemeValueUnknown() ContentTypeMetadataTaxonomyItemConceptSchemeValue {
	return ContentTypeMetadataTaxonomyItemConceptSchemeValue{
		state: attr.ValueStateUnknown,
	}
}

func NewContentTypeMetadataTaxonomyItemConceptSchemeValueNull() ContentTypeMetadataTaxonomyItemConceptSchemeValue {
	return ContentTypeMetadataTaxonomyItemConceptSchemeValue{
		state: attr.ValueStateNull,
	}
}

func NewContentTypeMetadataTaxonomyItemConceptSchemeValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (ContentTypeMetadataTaxonomyItemConceptSchemeValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := ContentTypeMetadataTaxonomyItemConceptSchemeValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

//nolint:ireturn
func (v ContentTypeMetadataTaxonomyItemConceptSchemeValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return ContentTypeMetadataTaxonomyItemConceptSchemeType{ObjectType: v.ObjectType(ctx)}
}

//nolint:ireturn
func (v ContentTypeMetadataTaxonomyItemConceptSchemeValue) Type(ctx context.Context) attr.Type {
	return ContentTypeMetadataTaxonomyItemConceptSchemeType{ObjectType: v.ObjectType(ctx)}
}

func (v ContentTypeMetadataTaxonomyItemConceptSchemeValue) ObjectType(ctx context.Context) types.ObjectType {
	return types.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v ContentTypeMetadataTaxonomyItemConceptSchemeValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v ContentTypeMetadataTaxonomyItemConceptSchemeValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[ContentTypeMetadataTaxonomyItemConceptSchemeValue](v, o)
}

func (v ContentTypeMetadataTaxonomyItemConceptSchemeValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ContentTypeMetadataTaxonomyItemConceptSchemeValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ContentTypeMetadataTaxonomyItemConceptSchemeValue) String() string {
	return "ContentTypeMetadataTaxonomyItemConceptSchemeValue"
}

func (v ContentTypeMetadataTaxonomyItemConceptSchemeValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v ContentTypeMetadataTaxonomyItemConceptSchemeValue) ToObjectValue(ctx context.Context) (types.Object, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
