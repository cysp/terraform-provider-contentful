package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (m ContentTypeMetadataValue) ToOptContentTypeMetadata(ctx context.Context, path path.Path) (cm.OptContentTypeMetadata, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if m.IsNull() {
		return cm.OptContentTypeMetadata{}, diags
	}

	taxonomy := ContentTypeMetadataTaxonomyItemsToContentTypeMetadataTaxonomySlice(ctx, path.AtName("taxonomy"), m.Taxonomy)

	metadata := cm.ContentTypeMetadata{
		Annotations: []byte(m.Annotations.ValueString()),
		Taxonomy:    taxonomy,
	}

	return cm.NewOptContentTypeMetadata(metadata), diags
}

func ContentTypeMetadataTaxonomyItemsToContentTypeMetadataTaxonomySlice(
	ctx context.Context,
	path path.Path,
	items TypedList[ContentTypeMetadataTaxonomyItemValue],
) []cm.ContentTypeMetadataTaxonomyItem {
	if items.IsNull() || items.IsUnknown() {
		return nil
	}

	itemValues := items.Elements()

	requestItems := make([]cm.ContentTypeMetadataTaxonomyItem, 0, len(itemValues))

	for index, itemValue := range itemValues {
		item := itemValue.ToContentTypeMetadataTaxonomyItem(ctx, path.AtListIndex(index))

		requestItems = append(requestItems, item...)
	}

	return requestItems
}

func (v ContentTypeMetadataTaxonomyItemValue) ToContentTypeMetadataTaxonomyItem(
	_ context.Context,
	_ path.Path,
) []cm.ContentTypeMetadataTaxonomyItem {
	if v.IsUnknown() || v.IsNull() {
		return nil
	}

	items := make([]cm.ContentTypeMetadataTaxonomyItem, 0, 1)

	if !v.TaxonomyConceptScheme.IsUnknown() && !v.TaxonomyConceptScheme.IsNull() {
		items = append(items, cm.ContentTypeMetadataTaxonomyItem{
			Sys: cm.ContentTypeMetadataTaxonomyItemSys{
				Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
				LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConceptScheme,
				ID:       v.TaxonomyConceptScheme.ID.ValueString(),
			},
			Required: cm.NewOptPointerBool(v.TaxonomyConceptScheme.Required.ValueBoolPointer()),
		})
	}

	if !v.TaxonomyConcept.IsUnknown() && !v.TaxonomyConcept.IsNull() {
		items = append(items, cm.ContentTypeMetadataTaxonomyItem{
			Sys: cm.ContentTypeMetadataTaxonomyItemSys{
				Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
				LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConcept,
				ID:       v.TaxonomyConcept.ID.ValueString(),
			},
			Required: cm.NewOptPointerBool(v.TaxonomyConcept.Required.ValueBoolPointer()),
		})
	}

	return items
}
