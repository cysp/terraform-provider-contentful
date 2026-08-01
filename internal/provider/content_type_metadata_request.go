package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func ToOptContentTypeMetadata(ctx context.Context, path path.Path, m TypedObject[ContentTypeMetadataValue]) (cm.OptContentTypeMetadata, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value, valueOk := m.GetValue()
	if !valueOk {
		return cm.OptContentTypeMetadata{}, diags
	}

	taxonomy, taxonomyDiags := ContentTypeMetadataTaxonomyItemsToContentTypeMetadataTaxonomySlice(ctx, path.AtName("taxonomy"), value.Taxonomy)
	diags.Append(taxonomyDiags...)

	if taxonomyDiags.HasError() {
		return cm.OptContentTypeMetadata{}, diags
	}

	var annotations []byte
	if !value.Annotations.IsNull() && !value.Annotations.IsUnknown() {
		annotations = []byte(value.Annotations.ValueString())
	}

	metadata := cm.ContentTypeMetadata{
		Annotations: annotations,
		Taxonomy:    taxonomy,
	}

	return cm.NewOptContentTypeMetadata(metadata), diags
}

func ContentTypeMetadataTaxonomyItemsToContentTypeMetadataTaxonomySlice(
	ctx context.Context,
	path path.Path,
	items TypedList[TypedObject[ContentTypeMetadataTaxonomyItemValue]],
) ([]cm.ContentTypeMetadataTaxonomyItem, diag.Diagnostics) {
	if items.IsNull() || items.IsUnknown() {
		return nil, nil
	}

	diags := diag.Diagnostics{}
	itemValues := items.Elements()

	requestItems := make([]cm.ContentTypeMetadataTaxonomyItem, 0, len(itemValues))

	for index, itemValue := range itemValues {
		item, itemDiags := ToContentTypeMetadataTaxonomyItem(ctx, path.AtListIndex(index), itemValue)
		diags.Append(itemDiags...)

		if !itemDiags.HasError() {
			requestItems = append(requestItems, item...)
		}
	}

	if diags.HasError() {
		return nil, diags
	}

	return requestItems, diags
}

func ToContentTypeMetadataTaxonomyItem(
	_ context.Context,
	path path.Path,
	object TypedObject[ContentTypeMetadataTaxonomyItemValue],
) ([]cm.ContentTypeMetadataTaxonomyItem, diag.Diagnostics) {
	value, valueDiags := KnownObjectValue(object, path)
	if valueDiags.HasError() {
		return nil, valueDiags
	}

	conceptPath := path.AtName("taxonomy_concept")
	conceptSchemePath := path.AtName("taxonomy_concept_scheme")
	item, diags := ConvertExactlyOneKnownAlternative(
		path,
		KnownUnionAlternative[cm.ContentTypeMetadataTaxonomyItem]{
			Name: "taxonomy_concept", Path: conceptPath, Value: value.TaxonomyConcept,
			Convert: func() (cm.ContentTypeMetadataTaxonomyItem, diag.Diagnostics) {
				taxonomyConcept, _ := value.TaxonomyConcept.GetValue()
				conceptID, conceptIDDiags := KnownStringValue(taxonomyConcept.ID, conceptPath.AtName("id"))
				required, requiredDiags := KnownBoolValue(taxonomyConcept.Required, conceptPath.AtName("required"))
				conversionDiags := diag.Diagnostics{}
				conversionDiags.Append(conceptIDDiags...)
				conversionDiags.Append(requiredDiags...)

				return cm.ContentTypeMetadataTaxonomyItem{
					Sys: cm.ContentTypeMetadataTaxonomyItemSys{
						Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
						LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConcept,
						ID:       conceptID,
					},
					Required: cm.NewOptBool(required),
				}, conversionDiags
			},
		},
		KnownUnionAlternative[cm.ContentTypeMetadataTaxonomyItem]{
			Name: "taxonomy_concept_scheme", Path: conceptSchemePath, Value: value.TaxonomyConceptScheme,
			Convert: func() (cm.ContentTypeMetadataTaxonomyItem, diag.Diagnostics) {
				taxonomyConceptScheme, _ := value.TaxonomyConceptScheme.GetValue()
				conceptSchemeID, conceptSchemeIDDiags := KnownStringValue(taxonomyConceptScheme.ID, conceptSchemePath.AtName("id"))
				required, requiredDiags := KnownBoolValue(taxonomyConceptScheme.Required, conceptSchemePath.AtName("required"))
				conversionDiags := diag.Diagnostics{}
				conversionDiags.Append(conceptSchemeIDDiags...)
				conversionDiags.Append(requiredDiags...)

				return cm.ContentTypeMetadataTaxonomyItem{
					Sys: cm.ContentTypeMetadataTaxonomyItemSys{
						Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
						LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConceptScheme,
						ID:       conceptSchemeID,
					},
					Required: cm.NewOptBool(required),
				}, conversionDiags
			},
		},
	)

	if diags.HasError() {
		return nil, diags
	}

	return []cm.ContentTypeMetadataTaxonomyItem{item}, diags
}
