package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ContentTypeMetadataValue struct {
	Annotations jsontypes.Normalized                                         `tfsdk:"annotations"`
	Taxonomy    TypedList[TypedObject[ContentTypeMetadataTaxonomyItemValue]] `tfsdk:"taxonomy"`
}

type ContentTypeMetadataTaxonomyItemValue struct {
	TaxonomyConcept       TypedObject[ContentTypeMetadataTaxonomyItemConceptValue]       `tfsdk:"taxonomy_concept"`
	TaxonomyConceptScheme TypedObject[ContentTypeMetadataTaxonomyItemConceptSchemeValue] `tfsdk:"taxonomy_concept_scheme"`
}

type ContentTypeMetadataTaxonomyItemConceptValue struct {
	ID       types.String `tfsdk:"id"`
	Required types.Bool   `tfsdk:"required"`
}

type ContentTypeMetadataTaxonomyItemConceptSchemeValue struct {
	ID       types.String `tfsdk:"id"`
	Required types.Bool   `tfsdk:"required"`
}
