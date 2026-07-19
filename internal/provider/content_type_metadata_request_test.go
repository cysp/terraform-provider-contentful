package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentTypeMetadataRequestSerialization(t *testing.T) {
	t.Parallel()

	taxonomyConceptScheme := NewTypedObject(ContentTypeMetadataTaxonomyItemConceptSchemeValue{
		ID:       types.StringValue("furniture"),
		Required: types.BoolValue(true),
	})

	tests := map[string]struct {
		metadata TypedObject[ContentTypeMetadataValue]
		expected string
	}{
		"absent metadata": {
			metadata: NewTypedObjectNull[ContentTypeMetadataValue](),
			expected: `{"name":"Test","description":null,"displayField":"title","fields":[]}`,
		},
		"annotations with taxonomy omitted": {
			metadata: NewTypedObject(ContentTypeMetadataValue{
				Annotations: NewNormalizedJSONTypesNormalizedValue([]byte(`{"ContentType":[]}`)),
				Taxonomy:    NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]](),
			}),
			expected: `{"name":"Test","description":null,"displayField":"title","fields":[],"metadata":{"annotations":{"ContentType":[]}}}`,
		},
		"empty taxonomy": {
			metadata: NewTypedObject(ContentTypeMetadataValue{
				Annotations: jsontypes.NewNormalizedNull(),
				Taxonomy:    NewTypedList([]TypedObject[ContentTypeMetadataTaxonomyItemValue]{}),
			}),
			expected: `{"name":"Test","description":null,"displayField":"title","fields":[],"metadata":{"taxonomy":[]}}`,
		},
		"populated taxonomy": {
			metadata: NewTypedObject(ContentTypeMetadataValue{
				Annotations: jsontypes.NewNormalizedNull(),
				Taxonomy: NewTypedList([]TypedObject[ContentTypeMetadataTaxonomyItemValue]{
					NewTypedObject(ContentTypeMetadataTaxonomyItemValue{
						TaxonomyConcept:       NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptValue](),
						TaxonomyConceptScheme: taxonomyConceptScheme,
					}),
				}),
			}),
			expected: `{"name":"Test","description":null,"displayField":"title","fields":[],"metadata":{"taxonomy":[{"sys":{"type":"Link","id":"furniture","linkType":"TaxonomyConceptScheme"},"required":true}]}}`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := ContentTypeModel{
				Name:         types.StringValue("Test"),
				DisplayField: types.StringValue("title"),
				Fields:       NewTypedList([]TypedObject[ContentTypeFieldValue]{}),
				Metadata:     test.metadata,
			}

			request, diags := model.ToContentTypeRequestData(t.Context())
			require.Empty(t, diags.Errors())

			requestBody, err := request.MarshalJSON()
			require.NoError(t, err)
			assert.JSONEq(t, test.expected, string(requestBody))
		})
	}
}

func TestContentTypeMetadataTaxonomyFailsWithoutPartialOutput(t *testing.T) {
	t.Parallel()

	items := NewTypedList([]TypedObject[ContentTypeMetadataTaxonomyItemValue]{
		NewTypedObject(ContentTypeMetadataTaxonomyItemValue{
			TaxonomyConcept: NewTypedObject(ContentTypeMetadataTaxonomyItemConceptValue{
				ID:       types.StringValue("valid"),
				Required: types.BoolValue(false),
			}),
			TaxonomyConceptScheme: NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptSchemeValue](),
		}),
		NewTypedObject(ContentTypeMetadataTaxonomyItemValue{
			TaxonomyConcept: NewTypedObject(ContentTypeMetadataTaxonomyItemConceptValue{
				ID:       types.StringUnknown(),
				Required: types.BoolValue(false),
			}),
			TaxonomyConceptScheme: NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptSchemeValue](),
		}),
	})

	result, diags := ContentTypeMetadataTaxonomyItemsToContentTypeMetadataTaxonomySlice(
		t.Context(),
		path.Root("metadata").AtName("taxonomy"),
		items,
	)

	assert.Nil(t, result)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{"metadata.taxonomy[1].taxonomy_concept.id"}, diagnosticPaths(t, diags))
}
