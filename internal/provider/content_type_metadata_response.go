package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewContentTypeMetadataFromResponse(path path.Path, optMetadata cm.OptContentTypeMetadata) (TypedObject[ContentTypeMetadataValue], diag.Diagnostics) {
	metadata, metadataOk := optMetadata.Get()
	if !metadataOk {
		return NewTypedObjectNull[ContentTypeMetadataValue](), nil
	}

	annotations := jsontypes.NewNormalizedNull()
	if metadata.Annotations != nil {
		annotations = NewNormalizedJSONValue(metadata.Annotations)
	}

	taxonomy, diags := NewContentTypeMetadataTaxonomyItemsFromResponse(path.AtName("taxonomy"), metadata.Taxonomy)

	return NewTypedObject(ContentTypeMetadataValue{
		Annotations: annotations,
		Taxonomy:    taxonomy,
	}), diags
}

func NewContentTypeMetadataTaxonomyItemsFromResponse(
	path path.Path,
	taxonomy []cm.ContentTypeMetadataTaxonomyItem,
) (TypedList[TypedObject[ContentTypeMetadataTaxonomyItemValue]], diag.Diagnostics) {
	if taxonomy == nil {
		return NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]](), diag.Diagnostics{}
	}

	diags := diag.Diagnostics{}

	items := make([]TypedObject[ContentTypeMetadataTaxonomyItemValue], 0, len(taxonomy))

	for index, item := range taxonomy {
		itemValue, itemValueDiags := NewContentTypeMetadataTaxonomyItemFromResponse(path.AtListIndex(index), item)
		diags.Append(itemValueDiags...)

		items = append(items, itemValue)
	}

	return NewTypedList(items), diags
}

func NewContentTypeMetadataTaxonomyItemFromResponse(
	itemPath path.Path,
	item cm.ContentTypeMetadataTaxonomyItem,
) (TypedObject[ContentTypeMetadataTaxonomyItemValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := ContentTypeMetadataTaxonomyItemValue{
		TaxonomyConcept:       NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptValue](),
		TaxonomyConceptScheme: NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptSchemeValue](),
	}

	switch item.Sys.LinkType {
	case cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConceptScheme:
		value.TaxonomyConceptScheme = NewTypedObject(ContentTypeMetadataTaxonomyItemConceptSchemeValue{
			ID:       types.StringValue(item.Sys.ID),
			Required: types.BoolPointerValue(item.Required.ValueBoolPointer()),
		})

	case cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConcept:
		value.TaxonomyConcept = NewTypedObject(ContentTypeMetadataTaxonomyItemConceptValue{
			ID:       types.StringValue(item.Sys.ID),
			Required: types.BoolPointerValue(item.Required.ValueBoolPointer()),
		})
	default:
		// Preserve an unsupported response discriminator as a known empty-union
		// sentinel so the list position and surrounding metadata remain visible.
		// Request conversion remains strict and rejects this sentinel before an
		// update.
		diags.AddAttributeWarning(itemPath, "Unsupported taxonomy metadata item", "Contentful returned an unknown taxonomy link type. Terraform state retains the item as an empty known union; a later request conversion will reject it until configured as a supported alternative.")
	}

	return NewTypedObject(value), diags
}
