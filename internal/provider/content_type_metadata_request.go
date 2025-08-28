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

	taxonomy := ContentTypeMetadataTaxonomyItemsToContentTypeMetadataTaxonomySlice(ctx, path.AtName("taxonomy"), value.Taxonomy)

	metadata := cm.ContentTypeMetadata{
		Annotations: []byte(value.Annotations.ValueString()),
		Taxonomy:    taxonomy,
	}

	return cm.NewOptContentTypeMetadata(metadata), diags
}

func ContentTypeMetadataTaxonomyItemsToContentTypeMetadataTaxonomySlice(
	ctx context.Context,
	path path.Path,
	items TypedList[TypedObject[ContentTypeMetadataTaxonomyItemValue]],
) []cm.ContentTypeMetadataTaxonomyItem {
	if items.IsNull() || items.IsUnknown() {
		return nil
	}

	itemValues := items.Elements()

	requestItems := make([]cm.ContentTypeMetadataTaxonomyItem, 0, len(itemValues))

	for index, itemValue := range itemValues {
		item := ToContentTypeMetadataTaxonomyItem(ctx, path.AtListIndex(index), itemValue)

		requestItems = append(requestItems, item...)
	}

	return requestItems
}

func ToContentTypeMetadataTaxonomyItem(
	_ context.Context,
	_ path.Path,
	object TypedObject[ContentTypeMetadataTaxonomyItemValue],
) []cm.ContentTypeMetadataTaxonomyItem {
	value, valueOk := object.GetValue()
	if !valueOk {
		return nil
	}

	items := make([]cm.ContentTypeMetadataTaxonomyItem, 0, 1)

	taxonomyConceptScheme, taxonomyConceptSchemeOk := value.TaxonomyConceptScheme.GetValue()
	if taxonomyConceptSchemeOk {
		items = append(items, cm.ContentTypeMetadataTaxonomyItem{
			Sys: cm.ContentTypeMetadataTaxonomyItemSys{
				Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
				LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConceptScheme,
				ID:       taxonomyConceptScheme.ID.ValueString(),
			},
			Required: cm.NewOptPointerBool(taxonomyConceptScheme.Required.ValueBoolPointer()),
		})
	}

	taxonomyConcept, taxonomyConceptOk := value.TaxonomyConcept.GetValue()
	if taxonomyConceptOk {
		items = append(items, cm.ContentTypeMetadataTaxonomyItem{
			Sys: cm.ContentTypeMetadataTaxonomyItemSys{
				Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
				LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConcept,
				ID:       taxonomyConcept.ID.ValueString(),
			},
			Required: cm.NewOptPointerBool(taxonomyConcept.Required.ValueBoolPointer()),
		})
	}

	return items
}
