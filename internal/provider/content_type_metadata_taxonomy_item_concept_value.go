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

type ContentTypeMetadataTaxonomyItemConceptValue struct {
	ID       types.String `tfsdk:"id"`
	Required types.Bool   `tfsdk:"required"`
	state    attr.ValueState
}

var _ basetypes.ObjectValuable = ContentTypeMetadataTaxonomyItemConceptValue{}

func NewContentTypeMetadataTaxonomyItemConceptValueUnknown() ContentTypeMetadataTaxonomyItemConceptValue {
	return ContentTypeMetadataTaxonomyItemConceptValue{
		state: attr.ValueStateUnknown,
	}
}

func NewContentTypeMetadataTaxonomyItemConceptValueNull() ContentTypeMetadataTaxonomyItemConceptValue {
	return ContentTypeMetadataTaxonomyItemConceptValue{
		state: attr.ValueStateNull,
	}
}

func NewContentTypeMetadataTaxonomyItemConceptValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (ContentTypeMetadataTaxonomyItemConceptValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := ContentTypeMetadataTaxonomyItemConceptValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

//nolint:ireturn
func (v ContentTypeMetadataTaxonomyItemConceptValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return ContentTypeMetadataTaxonomyItemConceptType{ObjectType: v.ObjectType(ctx)}
}

//nolint:ireturn
func (v ContentTypeMetadataTaxonomyItemConceptValue) Type(ctx context.Context) attr.Type {
	return ContentTypeMetadataTaxonomyItemConceptType{ObjectType: v.ObjectType(ctx)}
}

func (v ContentTypeMetadataTaxonomyItemConceptValue) ObjectType(ctx context.Context) types.ObjectType {
	return types.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v ContentTypeMetadataTaxonomyItemConceptValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v ContentTypeMetadataTaxonomyItemConceptValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[ContentTypeMetadataTaxonomyItemConceptValue](v, o)
}

func (v ContentTypeMetadataTaxonomyItemConceptValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ContentTypeMetadataTaxonomyItemConceptValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ContentTypeMetadataTaxonomyItemConceptValue) String() string {
	return "ContentTypeMetadataTaxonomyItemConceptValue"
}

func (v ContentTypeMetadataTaxonomyItemConceptValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v ContentTypeMetadataTaxonomyItemConceptValue) ToObjectValue(ctx context.Context) (types.Object, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
