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

		requestItems = append(requestItems, item...)
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

	diags := diag.Diagnostics{}
	items := make([]cm.ContentTypeMetadataTaxonomyItem, 0, 1)
	found := false

	taxonomyConceptScheme, taxonomyConceptSchemeOk := value.TaxonomyConceptScheme.GetValue()
	if taxonomyConceptSchemeOk {
		found = true
		conceptSchemeID, conceptSchemeIDDiags := KnownStringValue(taxonomyConceptScheme.ID, path.AtName("taxonomy_concept_scheme").AtName("id"))
		diags.Append(conceptSchemeIDDiags...)

		required, requiredDiags := KnownBoolValue(taxonomyConceptScheme.Required, path.AtName("taxonomy_concept_scheme").AtName("required"))
		diags.Append(requiredDiags...)

		items = append(items, cm.ContentTypeMetadataTaxonomyItem{
			Sys: cm.ContentTypeMetadataTaxonomyItemSys{
				Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
				LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConceptScheme,
				ID:       conceptSchemeID,
			},
			Required: cm.NewOptBool(required),
		})
	}

	taxonomyConcept, taxonomyConceptOk := value.TaxonomyConcept.GetValue()
	if taxonomyConceptOk {
		found = true
		conceptID, conceptIDDiags := KnownStringValue(taxonomyConcept.ID, path.AtName("taxonomy_concept").AtName("id"))
		diags.Append(conceptIDDiags...)

		required, requiredDiags := KnownBoolValue(taxonomyConcept.Required, path.AtName("taxonomy_concept").AtName("required"))
		diags.Append(requiredDiags...)

		items = append(items, cm.ContentTypeMetadataTaxonomyItem{
			Sys: cm.ContentTypeMetadataTaxonomyItemSys{
				Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
				LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConcept,
				ID:       conceptID,
			},
			Required: cm.NewOptBool(required),
		})
	}

	if !found {
		diags.AddAttributeError(path, "Missing content type taxonomy item", "Exactly one taxonomy concept or taxonomy concept scheme must be known and non-null.")
	}

	return items, diags
}
