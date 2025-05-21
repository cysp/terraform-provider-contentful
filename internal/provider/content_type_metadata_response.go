package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewContentTypeMetadataFromResponse(ctx context.Context, path path.Path, optMetadata cm.OptContentTypeMetadata) (ContentTypeMetadataValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	metadata, metadataOk := optMetadata.Get()
	if !metadataOk {
		return NewContentTypeMetadataValueNull(), diags
	}

	annotations := jsontypes.NewNormalizedNull()
	if metadata.Annotations != nil {
		annotations = jsontypes.NewNormalizedValue(string(metadata.Annotations))
	}

	taxonomy, taxonomyDiags := NewContentTypeMetadataTaxonomyItemsFromResponse(ctx, path.AtName("taxonomy"), metadata.Taxonomy)
	diags.Append(taxonomyDiags...)

	model, modelDiags := NewContentTypeMetadataValueKnownFromAttributes(ctx, map[string]attr.Value{
		"annotations": annotations,
		"taxonomy":    taxonomy,
	})
	diags.Append(modelDiags...)

	return model, diags
}

func NewContentTypeMetadataTaxonomyItemsFromResponse(
	ctx context.Context,
	path path.Path,
	taxonomy []cm.ContentTypeMetadataTaxonomyItem,
) (TypedList[ContentTypeMetadataTaxonomyItemValue], diag.Diagnostics) {
	if taxonomy == nil {
		return NewTypedListNull[ContentTypeMetadataTaxonomyItemValue](ctx), diag.Diagnostics{}
	}

	diags := diag.Diagnostics{}

	items := make([]ContentTypeMetadataTaxonomyItemValue, 0, len(taxonomy))

	for index, item := range taxonomy {
		itemValue, itemValueDiags := NewContentTypeMetadataTaxonomyItemFromResponse(ctx, path.AtListIndex(index), item)
		diags.Append(itemValueDiags...)

		items = append(items, itemValue)
	}

	list, listDiags := NewTypedList(ctx, items)
	diags.Append(listDiags...)

	return list, diags
}

func NewContentTypeMetadataTaxonomyItemFromResponse(
	ctx context.Context,
	_ path.Path,
	item cm.ContentTypeMetadataTaxonomyItem,
) (ContentTypeMetadataTaxonomyItemValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	attributes := make(map[string]attr.Value, 1)

	switch item.Sys.LinkType {
	case cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConceptScheme:
		value, valueDiags := NewContentTypeMetadataTaxonomyItemConceptSchemeValueKnownFromAttributes(ctx, map[string]attr.Value{
			"id":       types.StringValue(item.Sys.ID),
			"required": types.BoolPointerValue(item.Required.ValueBoolPointer()),
		})
		diags.Append(valueDiags...)

		attributes["taxonomy_concept_scheme"] = value
		attributes["taxonomy_concept"] = NewContentTypeMetadataTaxonomyItemConceptValueNull()

	case cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConcept:
		value, valueDiags := NewContentTypeMetadataTaxonomyItemConceptValueKnownFromAttributes(ctx, map[string]attr.Value{
			"id":       types.StringValue(item.Sys.ID),
			"required": types.BoolPointerValue(item.Required.ValueBoolPointer()),
		})
		diags.Append(valueDiags...)

		attributes["taxonomy_concept_scheme"] = NewContentTypeMetadataTaxonomyItemConceptSchemeValueNull()
		attributes["taxonomy_concept"] = value
	}

	return NewContentTypeMetadataTaxonomyItemValueKnownFromAttributes(ctx, attributes)
}
