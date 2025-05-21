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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type ContentTypeMetadataTaxonomyItemValue struct {
	TaxonomyConcept       ContentTypeMetadataTaxonomyItemConceptValue       `tfsdk:"taxonomy_concept"`
	TaxonomyConceptScheme ContentTypeMetadataTaxonomyItemConceptSchemeValue `tfsdk:"taxonomy_concept_scheme"`
	state                 attr.ValueState
}

var _ basetypes.ObjectValuable = ContentTypeMetadataTaxonomyItemValue{}

func NewContentTypeMetadataTaxonomyItemValueUnknown() ContentTypeMetadataTaxonomyItemValue {
	return ContentTypeMetadataTaxonomyItemValue{
		state: attr.ValueStateUnknown,
	}
}

func NewContentTypeMetadataTaxonomyItemValueNull() ContentTypeMetadataTaxonomyItemValue {
	return ContentTypeMetadataTaxonomyItemValue{
		state: attr.ValueStateNull,
	}
}

func NewContentTypeMetadataTaxonomyItemValueKnownFromAttributes(ctx context.Context, attributes map[string]attr.Value) (ContentTypeMetadataTaxonomyItemValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := ContentTypeMetadataTaxonomyItemValue{
		state: attr.ValueStateKnown,
	}

	setAttributesDiags := tpfr.SetAttributesInValue(ctx, &value, attributes)
	diags = append(diags, setAttributesDiags...)

	return value, diags
}

func (v ContentTypeMetadataTaxonomyItemValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"taxonomy_concept": schema.SingleNestedAttribute{
			Attributes: ContentTypeMetadataTaxonomyItemConceptValue{}.SchemaAttributes(ctx),
			CustomType: ContentTypeMetadataTaxonomyItemConceptValue{}.CustomType(ctx),
			Optional:   true,
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("taxonomy_concept"),
					path.MatchRelative().AtParent().AtName("taxonomy_concept_scheme"),
				),
			},
		},
		"taxonomy_concept_scheme": schema.SingleNestedAttribute{
			Attributes: ContentTypeMetadataTaxonomyItemConceptSchemeValue{}.SchemaAttributes(ctx),
			CustomType: ContentTypeMetadataTaxonomyItemConceptSchemeValue{}.CustomType(ctx),
			Optional:   true,
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("taxonomy_concept"),
					path.MatchRelative().AtParent().AtName("taxonomy_concept_scheme"),
				),
			},
		},
	}
}

//nolint:ireturn
func (v ContentTypeMetadataTaxonomyItemValue) CustomType(ctx context.Context) basetypes.ObjectTypable {
	return ContentTypeMetadataTaxonomyItemType{ObjectType: v.ObjectType(ctx)}
}

//nolint:ireturn
func (v ContentTypeMetadataTaxonomyItemValue) Type(ctx context.Context) attr.Type {
	return ContentTypeMetadataTaxonomyItemType{ObjectType: v.ObjectType(ctx)}
}

func (v ContentTypeMetadataTaxonomyItemValue) ObjectType(ctx context.Context) types.ObjectType {
	return types.ObjectType{AttrTypes: v.ObjectAttrTypes(ctx)}
}

func (v ContentTypeMetadataTaxonomyItemValue) ObjectAttrTypes(ctx context.Context) map[string]attr.Type {
	return ObjectAttrTypesFromSchemaAttributes(ctx, v.SchemaAttributes(ctx))
}

func (v ContentTypeMetadataTaxonomyItemValue) Equal(o attr.Value) bool {
	return tpfr.ValuesEqual[ContentTypeMetadataTaxonomyItemValue](v, o)
}

func (v ContentTypeMetadataTaxonomyItemValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v ContentTypeMetadataTaxonomyItemValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v ContentTypeMetadataTaxonomyItemValue) String() string {
	return "ContentTypeMetadataTaxonomyItemValue"
}

func (v ContentTypeMetadataTaxonomyItemValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	//nolint:wrapcheck
	return tpfr.ValueToTerraformValue(ctx, v, v.state)
}

func (v ContentTypeMetadataTaxonomyItemValue) ToObjectValue(ctx context.Context) (types.Object, diag.Diagnostics) {
	return tpfr.ValueToObjectValue(ctx, v)
}
