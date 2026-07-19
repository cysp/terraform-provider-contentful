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

	diags := diag.Diagnostics{}
	items := make([]cm.ContentTypeMetadataTaxonomyItem, 0, 1)
	found := 0

	taxonomyConceptScheme, taxonomyConceptSchemeOk := value.TaxonomyConceptScheme.GetValue()
	if taxonomyConceptSchemeOk {
		found++
		conceptSchemeID, conceptSchemeIDDiags := KnownStringValue(taxonomyConceptScheme.ID, path.AtName("taxonomy_concept_scheme").AtName("id"))
		diags.Append(conceptSchemeIDDiags...)

		required, requiredDiags := KnownBoolValue(taxonomyConceptScheme.Required, path.AtName("taxonomy_concept_scheme").AtName("required"))
		diags.Append(requiredDiags...)

		if !conceptSchemeIDDiags.HasError() && !requiredDiags.HasError() {
			items = append(items, cm.ContentTypeMetadataTaxonomyItem{
				Sys: cm.ContentTypeMetadataTaxonomyItemSys{
					Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
					LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConceptScheme,
					ID:       conceptSchemeID,
				},
				Required: cm.NewOptBool(required),
			})
		}
	}

	taxonomyConcept, taxonomyConceptOk := value.TaxonomyConcept.GetValue()
	if taxonomyConceptOk {
		found++
		conceptID, conceptIDDiags := KnownStringValue(taxonomyConcept.ID, path.AtName("taxonomy_concept").AtName("id"))
		diags.Append(conceptIDDiags...)

		required, requiredDiags := KnownBoolValue(taxonomyConcept.Required, path.AtName("taxonomy_concept").AtName("required"))
		diags.Append(requiredDiags...)

		if !conceptIDDiags.HasError() && !requiredDiags.HasError() {
			items = append(items, cm.ContentTypeMetadataTaxonomyItem{
				Sys: cm.ContentTypeMetadataTaxonomyItemSys{
					Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
					LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConcept,
					ID:       conceptID,
				},
				Required: cm.NewOptBool(required),
			})
		}
	}

	if found != 1 {
		diags.AddAttributeError(path, "Invalid content type taxonomy item", "Exactly one taxonomy concept or taxonomy concept scheme must be known and non-null.")
	}

	if diags.HasError() {
		return nil, diags
	}

	return items, diags
}
