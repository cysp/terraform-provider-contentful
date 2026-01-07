package testdata

import (
	"fmt"

	provider "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"pgregory.net/rapid"
)

func ContentTypeModel(spaceID, environmentID, contentTypeID string) *rapid.Generator[provider.ContentTypeModel] {
	return rapid.Custom(func(t *rapid.T) provider.ContentTypeModel {
		fieldIds := rapid.SliceOfNDistinct(AlphanumericStringOfN(1, 10), 0, 5, rapid.ID).Draw(t, "fieldIds")
		displayField := ""
		if len(fieldIds) > 0 {
			displayField = rapid.SampledFrom(fieldIds).Draw(t, "displayField")
		}

		return provider.ContentTypeModel{
			IDIdentityModel:          provider.NewIDIdentityModelFromMultipartID(spaceID, environmentID, contentTypeID),
			ContentTypeIdentityModel: provider.NewContentTypeIdentityModel(spaceID, environmentID, contentTypeID),
			Name:                     rapid.Map(AlphanumericStringOfN(1, 10), types.StringValue).Draw(t, "name"),
			Description:              rapid.Map(rapid.StringN(1, 20, 20), types.StringValue).Draw(t, "description"),
			DisplayField:             types.StringValue(displayField),
			Fields:                   ContentTypeFields(fieldIds).Draw(t, "fields"),
			Metadata:                 rapid.Map(ContentTypeMetadataValue(), provider.NewTypedObject).Draw(t, "metadata"),
		}
	})
}

func ContentTypeFields(fieldIds []string) *rapid.Generator[provider.TypedList[provider.TypedObject[provider.ContentTypeFieldValue]]] {
	return rapid.Custom(func(t *rapid.T) provider.TypedList[provider.TypedObject[provider.ContentTypeFieldValue]] {
		if len(fieldIds) == 0 {
			rapid.Bool().Draw(t, "empty")
			return provider.NewTypedList([]provider.TypedObject[provider.ContentTypeFieldValue]{})
		}

		values := make([]provider.TypedObject[provider.ContentTypeFieldValue], len(fieldIds))
		for i, fieldId := range fieldIds {
			values[i] = rapid.Map(ContentTypeFieldValue(fieldId), provider.NewTypedObject).Draw(t, fmt.Sprintf("fields[%s]", fieldId))
		}
		return provider.NewTypedList(values)
	})
}

func ContentTypeFieldValue(fieldId string) *rapid.Generator[provider.ContentTypeFieldValue] {
	return rapid.Custom(func(t *rapid.T) provider.ContentTypeFieldValue {
		return provider.ContentTypeFieldValue{
			ID:               types.StringValue(fieldId),
			Name:             rapid.Map(AlphanumericStringOfN(1, 10), types.StringValue).Draw(t, "name"),
			FieldType:        types.StringValue("Symbol"),
			LinkType:         RandomZeroable(rapid.Map(AlphanumericStringOfN(1, 10), types.StringValue)).Draw(t, "linkType"),
			Disabled:         rapid.Map(rapid.Bool(), types.BoolValue).Draw(t, "disabled"),
			Omitted:          rapid.Map(rapid.Bool(), types.BoolValue).Draw(t, "omitted"),
			Required:         rapid.Map(rapid.Bool(), types.BoolValue).Draw(t, "required"),
			DefaultValue:     RandomZeroable(JSONTypesNormalizedStringValue()).Draw(t, "defaultValue"),
			Items:            RandomZeroable(rapid.Map(ContentTypeFieldItemsValue(), provider.NewTypedObject)).Draw(t, "items"),
			Localized:        rapid.Map(rapid.Bool(), types.BoolValue).Draw(t, "localized"),
			Validations:      rapid.Map(rapid.SliceOfN(JSONTypesNormalizedStringValue(), 0, 2), provider.NewTypedList).Draw(t, "validations"),
			AllowedResources: RandomZeroable(rapid.Map(rapid.SliceOfN(rapid.Map(ContentTypeFieldAllowedResourceItemValue(), provider.NewTypedObject), 0, 2), provider.NewTypedList)).Draw(t, "allowedResources"),
		}
	})
}

func ContentTypeFieldItemsValue() *rapid.Generator[provider.ContentTypeFieldItemsValue] {
	return rapid.Custom(func(t *rapid.T) provider.ContentTypeFieldItemsValue {
		return provider.ContentTypeFieldItemsValue{
			ItemsType:   types.StringValue("Link"),
			LinkType:    types.StringValue("Entry"),
			Validations: rapid.Map(rapid.SliceOfN(JSONTypesNormalizedStringValue(), 0, 2), provider.NewTypedList).Draw(t, "validations"),
		}
	})
}

func ContentTypeFieldAllowedResourceItemValue() *rapid.Generator[provider.ContentTypeFieldAllowedResourceItemValue] {
	return rapid.OneOf(
		ContentTypeFieldAllowedResourceItemValueContentfulEntry(),
		ContentTypeFieldAllowedResourceItemValueExternal(),
	)
}

func ContentTypeFieldAllowedResourceItemValueContentfulEntry() *rapid.Generator[provider.ContentTypeFieldAllowedResourceItemValue] {
	return rapid.Custom(func(t *rapid.T) provider.ContentTypeFieldAllowedResourceItemValue {
		return provider.ContentTypeFieldAllowedResourceItemValue{
			ContentfulEntry: rapid.Map(ContentTypeFieldAllowedResourceItemContentfulEntryValue(), provider.NewTypedObject).Draw(t, "contentfulEntry"),
		}
	})
}

func ContentTypeFieldAllowedResourceItemValueExternal() *rapid.Generator[provider.ContentTypeFieldAllowedResourceItemValue] {
	return rapid.Custom(func(t *rapid.T) provider.ContentTypeFieldAllowedResourceItemValue {
		return provider.ContentTypeFieldAllowedResourceItemValue{
			External: rapid.Map(ContentTypeFieldAllowedResourceItemExternalValue(), provider.NewTypedObject).Draw(t, "external"),
		}
	})
}

func ContentTypeFieldAllowedResourceItemContentfulEntryValue() *rapid.Generator[provider.ContentTypeFieldAllowedResourceItemContentfulEntryValue] {
	return rapid.Custom(func(t *rapid.T) provider.ContentTypeFieldAllowedResourceItemContentfulEntryValue {
		return provider.ContentTypeFieldAllowedResourceItemContentfulEntryValue{
			Source:       rapid.Map(AlphanumericStringOfN(1, 10), types.StringValue).Draw(t, "source"),
			ContentTypes: rapid.Map(rapid.SliceOfN(AlphanumericStringOfN(1, 10), 0, 3), provider.NewTypedListFromStringSlice).Draw(t, "contentTypes"),
		}
	})
}

func ContentTypeFieldAllowedResourceItemExternalValue() *rapid.Generator[provider.ContentTypeFieldAllowedResourceItemExternalValue] {
	return rapid.Custom(func(t *rapid.T) provider.ContentTypeFieldAllowedResourceItemExternalValue {
		return provider.ContentTypeFieldAllowedResourceItemExternalValue{
			TypeID: rapid.Map(AlphanumericStringOfN(1, 10), types.StringValue).Draw(t, "typeID"),
		}
	})
}

func ContentTypeMetadataValue() *rapid.Generator[provider.ContentTypeMetadataValue] {
	return rapid.Custom(func(t *rapid.T) provider.ContentTypeMetadataValue {
		return provider.ContentTypeMetadataValue{
			Annotations: RandomZeroable(JSONTypesNormalizedStringValue()).Draw(t, "annotations"),
			Taxonomy:    RandomZeroable(ContentTypeMetadataTaxonomyList()).Draw(t, "taxonomy"),
		}
	})
}

func ContentTypeMetadataTaxonomyItemConceptValue() *rapid.Generator[provider.ContentTypeMetadataTaxonomyItemConceptValue] {
	return rapid.Custom(func(t *rapid.T) provider.ContentTypeMetadataTaxonomyItemConceptValue {
		return provider.ContentTypeMetadataTaxonomyItemConceptValue{
			ID:       rapid.Map(AlphanumericStringOfN(1, 10), types.StringValue).Draw(t, "id"),
			Required: rapid.Map(rapid.Bool(), types.BoolValue).Draw(t, "required"),
		}
	})
}

func ContentTypeMetadataTaxonomyItemConceptSchemeValue() *rapid.Generator[provider.ContentTypeMetadataTaxonomyItemConceptSchemeValue] {
	return rapid.Custom(func(t *rapid.T) provider.ContentTypeMetadataTaxonomyItemConceptSchemeValue {
		return provider.ContentTypeMetadataTaxonomyItemConceptSchemeValue{
			ID:       rapid.Map(AlphanumericStringOfN(1, 10), types.StringValue).Draw(t, "id"),
			Required: rapid.Map(rapid.Bool(), types.BoolValue).Draw(t, "required"),
		}
	})
}

func ContentTypeMetadataTaxonomyList() *rapid.Generator[provider.TypedList[provider.TypedObject[provider.ContentTypeMetadataTaxonomyItemValue]]] {
	return rapid.Map(
		rapid.SliceOfN(
			rapid.Map(ContentTypeMetadataTaxonomyItemValue(), provider.NewTypedObject),
			0,
			3,
		),
		provider.NewTypedList,
	)
}

func ContentTypeMetadataTaxonomyItemValue() *rapid.Generator[provider.ContentTypeMetadataTaxonomyItemValue] {
	return rapid.OneOf(
		ContentTypeMetadataTaxonomyItemValueConcept(),
		ContentTypeMetadataTaxonomyItemValueConceptScheme(),
	)
}

func ContentTypeMetadataTaxonomyItemValueConcept() *rapid.Generator[provider.ContentTypeMetadataTaxonomyItemValue] {
	return rapid.Custom(func(t *rapid.T) provider.ContentTypeMetadataTaxonomyItemValue {
		return provider.ContentTypeMetadataTaxonomyItemValue{
			TaxonomyConcept: rapid.Map(ContentTypeMetadataTaxonomyItemConceptValue(), provider.NewTypedObject).Draw(t, "taxonomyConcept"),
		}
	})
}

func ContentTypeMetadataTaxonomyItemValueConceptScheme() *rapid.Generator[provider.ContentTypeMetadataTaxonomyItemValue] {
	return rapid.Custom(func(t *rapid.T) provider.ContentTypeMetadataTaxonomyItemValue {
		return provider.ContentTypeMetadataTaxonomyItemValue{
			TaxonomyConceptScheme: rapid.Map(ContentTypeMetadataTaxonomyItemConceptSchemeValue(), provider.NewTypedObject).Draw(t, "taxonomyConceptScheme"),
		}
	})
}
