package provider_test

import (
	"fmt"
	"slices"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestContentTypeMutationReconciliationProperties(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	rapid.Check(t, func(t *rapid.T) {
		count := rapid.IntRange(2, 5).Draw(t, "field count")
		plannedFields := make([]TypedObject[ContentTypeFieldValue], count)
		returnedFields := make([]TypedObject[ContentTypeFieldValue], count)
		fields := make([]cm.ContentTypeFieldsItem, count)

		reversedDefaults := make([]string, count)
		for index := range count {
			plannedJSON, returnedJSON, reversedJSON := drawReconciliationJSON(t)
			reversedDefaults[index] = reversedJSON
			fieldID := fmt.Sprintf("field-%d", index)
			required := rapid.Bool().Draw(t, "required field")
			localized := rapid.Bool().Draw(t, "localized field")
			plannedFields[index] = NewTypedObject(ContentTypeFieldValue{
				ID: types.StringValue(fieldID), Name: types.StringValue(fieldID), FieldType: types.StringValue("Array"), LinkType: types.StringNull(),
				Disabled: types.BoolValue(false), Omitted: types.BoolValue(false), Required: types.BoolValue(required), Localized: types.BoolValue(localized),
				DefaultValue: jsontypes.NewNormalizedValue(plannedJSON),
				Items: NewTypedObject(ContentTypeFieldItemsValue{
					ItemsType: types.StringValue("Symbol"), LinkType: types.StringNull(),
					Validations: NewTypedList([]jsontypes.Normalized{jsontypes.NewNormalizedValue(plannedJSON), jsontypes.NewNormalizedValue(reversedJSON)}),
				}),
				Validations:      NewTypedList([]jsontypes.Normalized{jsontypes.NewNormalizedValue(plannedJSON), jsontypes.NewNormalizedValue(reversedJSON)}),
				AllowedResources: NewTypedListNull[TypedObject[ContentTypeFieldAllowedResourceItemValue]](),
			})
			returnedFields[index] = NewTypedObject(ContentTypeFieldValue{
				ID: types.StringValue(fieldID), Name: types.StringValue(fieldID), FieldType: types.StringValue("Array"), LinkType: types.StringNull(),
				Disabled: types.BoolValue(false), Omitted: types.BoolValue(false), Required: types.BoolValue(required), Localized: types.BoolValue(localized),
				DefaultValue: jsontypes.NewNormalizedValue(returnedJSON),
				Items: NewTypedObject(ContentTypeFieldItemsValue{
					ItemsType: types.StringValue("Symbol"), LinkType: types.StringNull(),
					Validations: NewTypedList([]jsontypes.Normalized{jsontypes.NewNormalizedValue(returnedJSON), jsontypes.NewNormalizedValue(reversedJSON)}),
				}),
				Validations:      NewTypedList([]jsontypes.Normalized{jsontypes.NewNormalizedValue(returnedJSON), jsontypes.NewNormalizedValue(reversedJSON)}),
				AllowedResources: NewTypedListNull[TypedObject[ContentTypeFieldAllowedResourceItemValue]](),
			})
			fields[index] = cm.ContentTypeFieldsItem{
				ID: fieldID, Name: fieldID, Type: "Array", Required: cm.NewOptBool(required), Localized: cm.NewOptBool(localized),
				Disabled: cm.NewOptBool(false), Omitted: cm.NewOptBool(false), DefaultValue: jx.Raw(returnedJSON),
				Items: cm.NewOptContentTypeFieldsItemItems(cm.ContentTypeFieldsItemItems{
					Type: cm.NewOptString("Symbol"), Validations: []jx.Raw{jx.Raw(returnedJSON), jx.Raw(reversedJSON)},
				}),
				Validations: []jx.Raw{jx.Raw(returnedJSON), jx.Raw(reversedJSON)},
			}
		}

		plannedJSON, returnedJSON, reversedJSON := drawReconciliationJSON(t)
		taxonomyRequired := rapid.Bool().Draw(t, "required taxonomy")
		taxonomy := NewTypedList([]TypedObject[ContentTypeMetadataTaxonomyItemValue]{NewTypedObject(ContentTypeMetadataTaxonomyItemValue{
			TaxonomyConcept: NewTypedObject(ContentTypeMetadataTaxonomyItemConceptValue{
				ID: types.StringValue("concept"), Required: types.BoolValue(taxonomyRequired),
			}),
			TaxonomyConceptScheme: NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptSchemeValue](),
		})})

		plan := ContentTypeModel{
			IDIdentityModel:          NewIDIdentityModelFromMultipartID("space", "master", "article"),
			ContentTypeIdentityModel: NewContentTypeIdentityModel("space", "master", "article"),
			Name:                     types.StringValue("Article"),
			Description:              types.StringValue("Description"),
			DisplayField:             types.StringValue("field-0"),
			PublishedVersion:         types.Int64Value(1000),
			Fields:                   NewTypedList(plannedFields),
			Metadata:                 NewTypedObject(ContentTypeMetadataValue{Annotations: jsontypes.NewNormalizedValue(plannedJSON), Taxonomy: taxonomy}),
			Timeouts:                 TimeoutsNull(),
		}
		if rapid.Bool().Draw(t, "response-owned taxonomy") {
			metadata := plan.Metadata.Value()
			metadata.Taxonomy = NewTypedListUnknown[TypedObject[ContentTypeMetadataTaxonomyItemValue]]()
			plan.Metadata = NewTypedObject(metadata)
		}

		response := cm.ContentType{
			Sys: cm.NewContentTypeSys("space", "master", "article"), Name: "Article", Description: cm.NewOptNilString("Description"),
			DisplayField: cm.NewNilString("field-0"), Fields: fields,
			Metadata: cm.NewOptContentTypeMetadata(cm.ContentTypeMetadata{
				Annotations: []byte(returnedJSON),
				Taxonomy: []cm.ContentTypeMetadataTaxonomyItem{{
					Sys:      cm.ContentTypeMetadataTaxonomyItemSys{Type: cm.ContentTypeMetadataTaxonomyItemSysTypeLink, LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConcept, ID: "concept"},
					Required: cm.NewOptBool(taxonomyRequired),
				}},
			}),
		}
		response.Sys.Version = rapid.IntRange(1, 100).Draw(t, "returned version")
		published := types.Int64Null()

		if rapid.Bool().Draw(t, "published response") {
			version := rapid.IntRange(0, response.Sys.Version-1).Draw(t, "published version")
			response.Sys.PublishedVersion.SetTo(version)
			published = types.Int64Value(int64(version))
		}

		expectedEquivalent := plan
		expectedEquivalent.PublishedVersion = published
		expectedEquivalent.Metadata = NewTypedObject(ContentTypeMetadataValue{Annotations: jsontypes.NewNormalizedValue(plannedJSON), Taxonomy: taxonomy})
		state, responseDiags, consistencyDiags, _ := ProjectContentTypeMutationResponse(ctx, response, plan)
		require.Empty(t, responseDiags)
		requireReconciliationPaths(t, consistencyDiags)
		assert.Equal(t, expectedEquivalent, state, "equivalent JSON restores exact Plan while unknown taxonomy and publication remain response-owned")

		expectedMetadata := NewTypedObject(ContentTypeMetadataValue{Annotations: jsontypes.NewNormalizedValue(returnedJSON), Taxonomy: taxonomy})
		index := rapid.IntRange(0, count-1).Draw(t, "changed field index")
		field := returnedFields[index].Value()

		var expectedPath string

		switch rapid.IntRange(0, 9).Draw(t, "contradiction") {
		case 0:
			field.Required = types.BoolValue(!field.Required.ValueBool())
			response.Fields[index].Required = cm.NewOptBool(field.Required.ValueBool())
			returnedFields[index] = NewTypedObject(field)
			expectedPath = fmt.Sprintf("fields[%d].required", index)
		case 1:
			field.DefaultValue = jsontypes.NewNormalizedValue(reversedDefaults[index])
			response.Fields[index].DefaultValue = jx.Raw(reversedDefaults[index])
			returnedFields[index] = NewTypedObject(field)
			expectedPath = fmt.Sprintf("fields[%d].default_value", index)
		case 2:
			items := field.Items.Value()
			items.Validations = NewTypedList([]jsontypes.Normalized{items.Validations.Elements()[1], items.Validations.Elements()[0]})
			field.Items = NewTypedObject(items)
			returnedFields[index] = NewTypedObject(field)
			validations := response.Fields[index].Items.Value.Validations
			validations[0], validations[1] = validations[1], validations[0]
			expectedPath = fmt.Sprintf("fields[%d].items.validations[0]", index)
		case 3:
			field.Validations = NewTypedList([]jsontypes.Normalized{field.Validations.Elements()[1], field.Validations.Elements()[0]})
			returnedFields[index] = NewTypedObject(field)
			response.Fields[index].Validations[0], response.Fields[index].Validations[1] = response.Fields[index].Validations[1], response.Fields[index].Validations[0]
			expectedPath = fmt.Sprintf("fields[%d].validations[0]", index)
		case 4:
			next := (index + 1) % count
			response.Fields[index], response.Fields[next] = response.Fields[next], response.Fields[index]
			returnedFields[index], returnedFields[next] = returnedFields[next], returnedFields[index]
			expectedPath = fmt.Sprintf("fields[%d].id", min(index, next))
		case 5:
			response.Metadata.Value.Annotations = []byte(reversedJSON)
			expectedMetadata = NewTypedObject(ContentTypeMetadataValue{Annotations: jsontypes.NewNormalizedValue(reversedJSON), Taxonomy: taxonomy})
			expectedPath = "metadata.annotations"
		case 6:
			plan.Metadata = NewTypedObject(ContentTypeMetadataValue{Annotations: jsontypes.NewNormalizedValue(plannedJSON), Taxonomy: taxonomy})
			response.Metadata.Value.Taxonomy[0].Required = cm.NewOptBool(!taxonomyRequired)
			changed := taxonomy.Elements()[0].Value()
			changed.TaxonomyConcept = NewTypedObject(ContentTypeMetadataTaxonomyItemConceptValue{ID: types.StringValue("concept"), Required: types.BoolValue(!taxonomyRequired)})
			expectedMetadata = NewTypedObject(ContentTypeMetadataValue{
				Annotations: jsontypes.NewNormalizedValue(returnedJSON), Taxonomy: NewTypedList([]TypedObject[ContentTypeMetadataTaxonomyItemValue]{NewTypedObject(changed)}),
			})
			expectedPath = "metadata.taxonomy"
		case 7:
			response.Sys.Environment.Sys.ID = "other-environment"
			expectedPath = "environment_id"
		case 8:
			plan.Metadata = NewTypedObject(ContentTypeMetadataValue{
				Annotations: jsontypes.NewNormalizedValue(plannedJSON), Taxonomy: NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]](),
			})
			response.Metadata.Value.Taxonomy = []cm.ContentTypeMetadataTaxonomyItem{}
			expectedMetadata = NewTypedObject(ContentTypeMetadataValue{
				Annotations: jsontypes.NewNormalizedValue(returnedJSON), Taxonomy: NewTypedList([]TypedObject[ContentTypeMetadataTaxonomyItemValue]{}),
			})
			expectedPath = "metadata.taxonomy"
		case 9:
			response.Fields = slices.Delete(slices.Clone(response.Fields), index, index+1)

			returnedFields = slices.Delete(slices.Clone(returnedFields), index, index+1)
			if index == count-1 {
				expectedPath = fmt.Sprintf("fields[%d]", index)
			} else {
				expectedPath = fmt.Sprintf("fields[%d].id", index)
			}
		}

		expected := ContentTypeModel{
			IDIdentityModel:          plan.IDIdentityModel,
			ContentTypeIdentityModel: plan.ContentTypeIdentityModel,
			Name:                     plan.Name,
			Description:              plan.Description,
			DisplayField:             plan.DisplayField,
			PublishedVersion:         published,
			Fields:                   NewTypedList(returnedFields),
			Metadata:                 expectedMetadata,
			Timeouts:                 plan.Timeouts,
		}

		state, responseDiags, consistencyDiags, _ = ProjectContentTypeMutationResponse(ctx, response, plan)
		require.Empty(t, responseDiags)
		requireReconciliationPaths(t, consistencyDiags, expectedPath)
		assert.Equal(t, expected, state, "nested contradiction must retain complete response, publication observation, and endpoint identity")
	})
}
