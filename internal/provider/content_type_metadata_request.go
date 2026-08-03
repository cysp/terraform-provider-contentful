package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToOptContentTypeMetadata(
	ctx context.Context,
	valuePath path.Path,
	metadataObject TypedObject[ContentTypeMetadataValue],
) (cm.OptContentTypeMetadata, diag.Diagnostics) {
	value, ok := metadataObject.GetValue()
	if !ok {
		// Optional+Computed metadata uses both null and unknown to represent omission
		// while plan reconciliation preserves remote taxonomy configuration.
		return cm.OptContentTypeMetadata{}, nil
	}

	diags := diag.Diagnostics{}
	taxonomy, taxonomyDiags := ContentTypeMetadataTaxonomyItemsToContentTypeMetadataTaxonomySlice(
		ctx,
		valuePath.AtName("taxonomy"),
		value.Taxonomy,
	)
	diags.Append(taxonomyDiags...)

	var annotations []byte

	if value.Annotations.IsUnknown() {
		diags.AddAttributeError(
			valuePath.AtName("annotations"),
			"Unexpected unknown content type metadata annotations",
			"Content type metadata annotations must be known before they can be sent to Contentful.",
		)
	} else if !value.Annotations.IsNull() {
		annotations = []byte(value.Annotations.ValueString())
	}

	if diags.HasError() {
		return cm.OptContentTypeMetadata{}, diags
	}

	return cm.NewOptContentTypeMetadata(cm.ContentTypeMetadata{
		Annotations: annotations,
		Taxonomy:    taxonomy,
	}), diags
}

func ContentTypeMetadataTaxonomyItemsToContentTypeMetadataTaxonomySlice(
	ctx context.Context,
	valuePath path.Path,
	items TypedList[TypedObject[ContentTypeMetadataTaxonomyItemValue]],
) ([]cm.ContentTypeMetadataTaxonomyItem, diag.Diagnostics) {
	if items.IsNull() || items.IsUnknown() {
		// taxonomy is Optional+Computed: an unresolved container represents
		// omission, while a known empty list explicitly removes taxonomy items.
		return nil, nil
	}

	return convertKnownObjectListElements(
		ctx,
		valuePath,
		items.Elements(),
		func(ctx context.Context, itemPath path.Path, value ContentTypeMetadataTaxonomyItemValue) (cm.ContentTypeMetadataTaxonomyItem, diag.Diagnostics) {
			return value.ToContentTypeMetadataTaxonomyItem(ctx, itemPath)
		},
	)
}

func (value ContentTypeMetadataTaxonomyItemValue) ToContentTypeMetadataTaxonomyItem(
	_ context.Context,
	valuePath path.Path,
) (cm.ContentTypeMetadataTaxonomyItem, diag.Diagnostics) {
	conceptPath := valuePath.AtName("taxonomy_concept")
	conceptSchemePath := valuePath.AtName("taxonomy_concept_scheme")

	return convertExactlyOneKnownAlternative(
		valuePath,
		knownUnionAlternative[cm.ContentTypeMetadataTaxonomyItem]{
			Name:  "taxonomy_concept",
			Path:  conceptPath,
			Value: value.TaxonomyConcept,
			Convert: func() (cm.ContentTypeMetadataTaxonomyItem, diag.Diagnostics) {
				concept, _ := value.TaxonomyConcept.GetValue()

				return contentTypeMetadataTaxonomyItem(
					concept.ID,
					concept.Required,
					conceptPath,
					cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConcept,
				)
			},
		},
		knownUnionAlternative[cm.ContentTypeMetadataTaxonomyItem]{
			Name:  "taxonomy_concept_scheme",
			Path:  conceptSchemePath,
			Value: value.TaxonomyConceptScheme,
			Convert: func() (cm.ContentTypeMetadataTaxonomyItem, diag.Diagnostics) {
				conceptScheme, _ := value.TaxonomyConceptScheme.GetValue()

				return contentTypeMetadataTaxonomyItem(
					conceptScheme.ID,
					conceptScheme.Required,
					conceptSchemePath,
					cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConceptScheme,
				)
			},
		},
	)
}

func contentTypeMetadataTaxonomyItem(
	idValue types.String,
	requiredValue types.Bool,
	valuePath path.Path,
	linkType cm.ContentTypeMetadataTaxonomyItemSysLinkType,
) (cm.ContentTypeMetadataTaxonomyItem, diag.Diagnostics) {
	taxonomyID, idDiags := requestRequiredString(idValue, valuePath.AtName("id"))
	required, requiredDiags := requestOptionalBool(requiredValue, valuePath.AtName("required"))
	diags := diag.Diagnostics{}
	diags.Append(idDiags...)
	diags.Append(requiredDiags...)

	if diags.HasError() {
		return cm.ContentTypeMetadataTaxonomyItem{}, diags
	}

	return cm.ContentTypeMetadataTaxonomyItem{
		Sys: cm.ContentTypeMetadataTaxonomyItemSys{
			Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
			LinkType: linkType,
			ID:       taxonomyID,
		},
		Required: required,
	}, diags
}
