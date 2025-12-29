package provider_test

import (
	"context"
	"fmt"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"pgregory.net/rapid"
)

func TestContentTypeModelRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	rapid.Check(t, func(t *rapid.T) {
		model := rapidContentTypeModelGenerator().Draw(t, "model")

		request, diags := model.ToContentTypeRequestData(ctx)
		if diags.HasError() {
			t.Fatalf("ToContentTypeRequestData failed: %v", diags.Errors())
		}

		contentType := cmt.NewContentTypeFromRequestFields(
			model.SpaceID.ValueString(),
			model.EnvironmentID.ValueString(),
			model.ContentTypeID.ValueString(),
			request,
		)

		result, diags := NewContentTypeResourceModelFromResponse(ctx, contentType)
		if diags.HasError() {
			t.Fatalf("NewContentTypeResourceModelFromResponse failed: %v", diags.Errors())
		}

		assert.Equal(t, model, result)
	})
}

func rapidContentTypeModelGenerator() *rapid.Generator[ContentTypeModel] {
	return rapid.Custom(func(t *rapid.T) ContentTypeModel {
		spaceID := rapid.StringMatching(`[a-zA-Z0-9]{1,10}`).Draw(t, "spaceID")
		environmentID := rapid.StringMatching(`[a-zA-Z0-9]{1,10}`).Draw(t, "environmentID")
		contentTypeID := rapid.StringMatching(`[a-zA-Z0-9]{1,10}`).Draw(t, "contentTypeID")

		name := rapid.StringMatching(`[a-zA-Z0-9]{1,20}`).Draw(t, "contentTypeName")
		description := rapid.StringMatching(`[a-zA-Z0-9]{0,40}`).Draw(t, "contentTypeDescription")
		displayField := rapid.StringMatching(`[a-zA-Z0-9]{0,20}`).Draw(t, "contentTypeDisplayField")

		fields := rapidContentTypeFields(t)
		metadata := rapidContentTypeMetadata(t)

		return ContentTypeModel{
			IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, environmentID, contentTypeID),
			ContentTypeIdentityModel: ContentTypeIdentityModel{
				SpaceID:       types.StringValue(spaceID),
				EnvironmentID: types.StringValue(environmentID),
				ContentTypeID: types.StringValue(contentTypeID),
			},
			Name:         types.StringValue(name),
			Description:  types.StringValue(description),
			DisplayField: types.StringValue(displayField),
			Fields:       fields,
			Metadata:     metadata,
		}
	})
}

func rapidContentTypeFields(t *rapid.T) TypedList[TypedObject[ContentTypeFieldValue]] {
	fieldCount := rapid.IntRange(0, 3).Draw(t, "fieldCount")
	fieldValues := make([]TypedObject[ContentTypeFieldValue], fieldCount)

	for i := range fieldCount {
		fieldValues[i] = rapidContentTypeField(t, i)
	}

	return NewTypedList(fieldValues)
}

func rapidContentTypeField(t *rapid.T, index int) TypedObject[ContentTypeFieldValue] {
	fieldID := rapid.StringMatching(`[a-zA-Z0-9]{1,10}`).Draw(t, fmt.Sprintf("fieldID[%d]", index))
	fieldName := rapid.StringMatching(`[a-zA-Z0-9]{1,20}`).Draw(t, fmt.Sprintf("fieldName[%d]", index))
	fieldType := rapid.SampledFrom([]string{"Symbol", "Text", "Link", "Array"}).Draw(t, fmt.Sprintf("fieldType[%d]", index))

	defaultValue := jsontypes.NewNormalizedNull()

	if rapid.Bool().Draw(t, fmt.Sprintf("hasDefaultValue[%d]", index)) {
		defaultToken := rapid.StringMatching(`[a-zA-Z0-9]{0,10}`).Draw(t, fmt.Sprintf("defaultToken[%d]", index))
		defaultValue = jsontypes.NewNormalizedValue(fmt.Sprintf(`"%s"`, defaultToken))
	}

	field := ContentTypeFieldValue{
		ID:               types.StringValue(fieldID),
		Name:             types.StringValue(fieldName),
		FieldType:        types.StringValue(fieldType),
		LinkType:         types.StringNull(),
		Disabled:         types.BoolValue(rapid.Bool().Draw(t, fmt.Sprintf("fieldDisabled[%d]", index))),
		Omitted:          types.BoolValue(rapid.Bool().Draw(t, fmt.Sprintf("fieldOmitted[%d]", index))),
		Required:         types.BoolValue(rapid.Bool().Draw(t, fmt.Sprintf("fieldRequired[%d]", index))),
		DefaultValue:     defaultValue,
		Items:            NewTypedObjectNull[ContentTypeFieldItemsValue](),
		Localized:        types.BoolValue(rapid.Bool().Draw(t, fmt.Sprintf("fieldLocalized[%d]", index))),
		Validations:      NewTypedList(rapidContentTypeValidations(t, index)),
		AllowedResources: NewTypedListNull[TypedObject[ContentTypeFieldAllowedResourceItemValue]](),
	}

	if fieldType == "Link" {
		field.LinkType = types.StringValue(rapid.SampledFrom([]string{"Entry", "Asset"}).Draw(t, fmt.Sprintf("fieldLinkType[%d]", index)))
	}

	if fieldType == "Array" {
		field.Items = rapidContentTypeItems(t, index)
	}

	if rapid.Bool().Draw(t, fmt.Sprintf("hasAllowedResources[%d]", index)) {
		field.AllowedResources = rapidAllowedResources(t, index)
	}

	return NewTypedObject(field)
}

func rapidContentTypeValidations(t *rapid.T, index int) []jsontypes.Normalized {
	validationCount := rapid.IntRange(0, 3).Draw(t, fmt.Sprintf("validationCount[%d]", index))
	validations := make([]jsontypes.Normalized, validationCount)

	for i := range validationCount {
		rule := rapid.StringMatching(`[a-z]{1,6}`).Draw(t, fmt.Sprintf("validationRule[%d][%d]", index, i))
		validations[i] = jsontypes.NewNormalizedValue(fmt.Sprintf(`{"rule":"%s"}`, rule))
	}

	return validations
}

func rapidContentTypeItems(t *rapid.T, index int) TypedObject[ContentTypeFieldItemsValue] {
	itemsType := rapid.SampledFrom([]string{"Link", "Symbol"}).Draw(t, fmt.Sprintf("itemsType[%d]", index))

	items := ContentTypeFieldItemsValue{
		ItemsType:   types.StringValue(itemsType),
		LinkType:    types.StringNull(),
		Validations: NewTypedList(rapidContentTypeItemValidations(t, index)),
	}

	if itemsType == "Link" {
		items.LinkType = types.StringValue(rapid.SampledFrom([]string{"Entry", "Asset"}).Draw(t, fmt.Sprintf("itemsLinkType[%d]", index)))
	}

	return NewTypedObject(items)
}

func rapidContentTypeItemValidations(t *rapid.T, index int) []jsontypes.Normalized {
	validationCount := rapid.IntRange(0, 2).Draw(t, fmt.Sprintf("itemValidationCount[%d]", index))
	validations := make([]jsontypes.Normalized, validationCount)

	for i := range validationCount {
		token := rapid.StringMatching(`[a-z]{1,6}`).Draw(t, fmt.Sprintf("itemValidationRule[%d][%d]", index, i))
		validations[i] = jsontypes.NewNormalizedValue(fmt.Sprintf(`{"item":"%s"}`, token))
	}

	return validations
}

func rapidAllowedResources(t *rapid.T, index int) TypedList[TypedObject[ContentTypeFieldAllowedResourceItemValue]] {
	resourceCount := rapid.IntRange(1, 2).Draw(t, fmt.Sprintf("allowedResourceCount[%d]", index))
	resources := make([]TypedObject[ContentTypeFieldAllowedResourceItemValue], resourceCount)

	for i := range resourceCount {
		if rapid.Bool().Draw(t, fmt.Sprintf("allowedResourceIsEntry[%d][%d]", index, i)) {
			contentTypes := rapid.SliceOfN(rapid.StringMatching(`[a-zA-Z0-9]{1,10}`), 0, 3).Draw(t, fmt.Sprintf("allowedResourceContentTypes[%d][%d]", index, i))

			resources[i] = NewTypedObject(ContentTypeFieldAllowedResourceItemValue{
				ContentfulEntry: NewTypedObject(ContentTypeFieldAllowedResourceItemContentfulEntryValue{
					Source:       types.StringValue(rapid.StringMatching(`[a-zA-Z0-9]{1,10}`).Draw(t, fmt.Sprintf("allowedResourceSource[%d][%d]", index, i))),
					ContentTypes: NewTypedListFromStringSlice(contentTypes),
				}),
				External: NewTypedObjectNull[ContentTypeFieldAllowedResourceItemExternalValue](),
			})

			continue
		}

		resources[i] = NewTypedObject(ContentTypeFieldAllowedResourceItemValue{
			ContentfulEntry: NewTypedObjectNull[ContentTypeFieldAllowedResourceItemContentfulEntryValue](),
			External: NewTypedObject(ContentTypeFieldAllowedResourceItemExternalValue{
				TypeID: types.StringValue(rapid.StringMatching(`[a-zA-Z0-9]{1,10}`).Draw(t, fmt.Sprintf("allowedResourceExternalType[%d][%d]", index, i))),
			}),
		})
	}

	return NewTypedList(resources)
}

func rapidContentTypeMetadata(t *rapid.T) TypedObject[ContentTypeMetadataValue] {
	if !rapid.Bool().Draw(t, "hasMetadata") {
		return NewTypedObjectNull[ContentTypeMetadataValue]()
	}

	annotationToken := rapid.StringMatching(`[a-zA-Z0-9]{1,10}`).Draw(t, "metadataAnnotationToken")

	return NewTypedObject(ContentTypeMetadataValue{
		Annotations: jsontypes.NewNormalizedValue(fmt.Sprintf(`{"note":"%s"}`, annotationToken)),
		Taxonomy:    rapidContentTypeMetadataTaxonomy(t),
	})
}

func rapidContentTypeMetadataTaxonomy(t *rapid.T) TypedList[TypedObject[ContentTypeMetadataTaxonomyItemValue]] {
	if !rapid.Bool().Draw(t, "hasMetadataTaxonomy") {
		return NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]]()
	}

	itemCount := rapid.IntRange(0, 3).Draw(t, "metadataTaxonomyCount")
	items := make([]TypedObject[ContentTypeMetadataTaxonomyItemValue], itemCount)

	for i := range itemCount {
		if rapid.Bool().Draw(t, fmt.Sprintf("taxonomyItemIsConcept[%d]", i)) {
			items[i] = NewTypedObject(ContentTypeMetadataTaxonomyItemValue{
				TaxonomyConcept:       rapidTaxonomyConcept(t, i),
				TaxonomyConceptScheme: NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptSchemeValue](),
			})

			continue
		}

		items[i] = NewTypedObject(ContentTypeMetadataTaxonomyItemValue{
			TaxonomyConcept:       NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptValue](),
			TaxonomyConceptScheme: rapidTaxonomyConceptScheme(t, i),
		})
	}

	return NewTypedList(items)
}

func rapidTaxonomyConcept(t *rapid.T, index int) TypedObject[ContentTypeMetadataTaxonomyItemConceptValue] {
	return NewTypedObject(ContentTypeMetadataTaxonomyItemConceptValue{
		ID:       types.StringValue(rapid.StringMatching(`[a-zA-Z0-9]{1,10}`).Draw(t, fmt.Sprintf("taxonomyConceptID[%d]", index))),
		Required: types.BoolValue(rapid.Bool().Draw(t, fmt.Sprintf("taxonomyConceptRequired[%d]", index))),
	})
}

func rapidTaxonomyConceptScheme(t *rapid.T, index int) TypedObject[ContentTypeMetadataTaxonomyItemConceptSchemeValue] {
	return NewTypedObject(ContentTypeMetadataTaxonomyItemConceptSchemeValue{
		ID:       types.StringValue(rapid.StringMatching(`[a-zA-Z0-9]{1,10}`).Draw(t, fmt.Sprintf("taxonomyConceptSchemeID[%d]", index))),
		Required: types.BoolValue(rapid.Bool().Draw(t, fmt.Sprintf("taxonomyConceptSchemeRequired[%d]", index))),
	})
}
