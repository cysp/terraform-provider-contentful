package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
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
		"unresolved computed metadata": {
			metadata: NewTypedObjectUnknown[ContentTypeMetadataValue](),
			expected: `{"name":"Test","description":null,"displayField":"title","fields":[]}`,
		},
		"annotations with taxonomy omitted": {
			metadata: NewTypedObject(ContentTypeMetadataValue{
				Annotations: NewNormalizedJSONValue([]byte(`{"ContentType":[]}`)),
				Taxonomy:    NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]](),
			}),
			expected: `{"name":"Test","description":null,"displayField":"title","fields":[],"metadata":{"annotations":{"ContentType":[]}}}`,
		},
		"annotations with unresolved computed taxonomy": {
			metadata: NewTypedObject(ContentTypeMetadataValue{
				Annotations: NewNormalizedJSONValue([]byte(`{"ContentType":[]}`)),
				Taxonomy:    NewTypedListUnknown[TypedObject[ContentTypeMetadataTaxonomyItemValue]](),
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

			request, diags := model.ToContentTypeRequestData(t.Context(), ContentTypeModel{})
			require.Empty(t, diags.Errors())

			requestBody, err := request.MarshalJSON()
			require.NoError(t, err)
			assert.JSONEq(t, test.expected, string(requestBody))
		})
	}
}

func TestContentTypeMetadataRejectsUnknownAnnotations(t *testing.T) {
	t.Parallel()

	actual, diags := ToOptContentTypeMetadata(
		t.Context(),
		path.Root("metadata"),
		NewTypedObject(ContentTypeMetadataValue{
			Annotations: jsontypes.NewNormalizedUnknown(),
			Taxonomy:    NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]](),
		}),
	)

	assert.Equal(t, cm.OptContentTypeMetadata{}, actual)
	assert.Equal(t, []string{"metadata.annotations"}, attributeDiagnosticPaths(t, diags))
}

func TestContentTypeMetadataTaxonomyRejectsUnresolvedElements(t *testing.T) {
	t.Parallel()

	items := NewTypedList([]TypedObject[ContentTypeMetadataTaxonomyItemValue]{
		NewTypedObject(ContentTypeMetadataTaxonomyItemValue{
			TaxonomyConcept: NewTypedObject(ContentTypeMetadataTaxonomyItemConceptValue{
				ID:       types.StringValue("valid"),
				Required: types.BoolValue(false),
			}),
			TaxonomyConceptScheme: NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptSchemeValue](),
		}),
		NewTypedObjectNull[ContentTypeMetadataTaxonomyItemValue](),
		NewTypedObjectUnknown[ContentTypeMetadataTaxonomyItemValue](),
	})

	result, diags := ContentTypeMetadataTaxonomyItemsToContentTypeMetadataTaxonomySlice(
		t.Context(),
		path.Root("metadata").AtName("taxonomy"),
		items,
	)

	assert.Nil(t, result)
	assert.Equal(t, []string{"metadata.taxonomy[1]", "metadata.taxonomy[2]"}, attributeDiagnosticPaths(t, diags))
}

func TestContentTypeMetadataTaxonomyRequiresExactlyOneAlternative(t *testing.T) {
	t.Parallel()

	valuePath := path.Root("metadata").AtName("taxonomy").AtListIndex(0)
	concept := NewTypedObject(ContentTypeMetadataTaxonomyItemConceptValue{
		ID:       types.StringValue("concept"),
		Required: types.BoolValue(false),
	})
	conceptScheme := NewTypedObject(ContentTypeMetadataTaxonomyItemConceptSchemeValue{
		ID:       types.StringValue("scheme"),
		Required: types.BoolValue(true),
	})

	tests := map[string]struct {
		value         ContentTypeMetadataTaxonomyItemValue
		expected      cm.ContentTypeMetadataTaxonomyItem
		expectedPaths []string
	}{
		"concept": {
			value: ContentTypeMetadataTaxonomyItemValue{
				TaxonomyConcept:       concept,
				TaxonomyConceptScheme: NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptSchemeValue](),
			},
			expected: cm.ContentTypeMetadataTaxonomyItem{
				Sys: cm.ContentTypeMetadataTaxonomyItemSys{
					Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
					LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConcept,
					ID:       "concept",
				},
				Required: cm.NewOptBool(false),
			},
			expectedPaths: []string{},
		},
		"concept scheme": {
			value: ContentTypeMetadataTaxonomyItemValue{
				TaxonomyConcept:       NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptValue](),
				TaxonomyConceptScheme: conceptScheme,
			},
			expected: cm.ContentTypeMetadataTaxonomyItem{
				Sys: cm.ContentTypeMetadataTaxonomyItemSys{
					Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
					LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConceptScheme,
					ID:       "scheme",
				},
				Required: cm.NewOptBool(true),
			},
			expectedPaths: []string{},
		},
		"neither": {
			value: ContentTypeMetadataTaxonomyItemValue{
				TaxonomyConcept:       NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptValue](),
				TaxonomyConceptScheme: NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptSchemeValue](),
			},
			expectedPaths: []string{"metadata.taxonomy[0]"},
		},
		"both": {
			value: ContentTypeMetadataTaxonomyItemValue{
				TaxonomyConcept:       concept,
				TaxonomyConceptScheme: conceptScheme,
			},
			expectedPaths: []string{"metadata.taxonomy[0]"},
		},
		"unknown concept": {
			value: ContentTypeMetadataTaxonomyItemValue{
				TaxonomyConcept:       NewTypedObjectUnknown[ContentTypeMetadataTaxonomyItemConceptValue](),
				TaxonomyConceptScheme: NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptSchemeValue](),
			},
			expectedPaths: []string{"metadata.taxonomy[0].taxonomy_concept"},
		},
		"unknown id": {
			value: ContentTypeMetadataTaxonomyItemValue{
				TaxonomyConcept: NewTypedObject(ContentTypeMetadataTaxonomyItemConceptValue{
					ID:       types.StringUnknown(),
					Required: types.BoolValue(false),
				}),
				TaxonomyConceptScheme: NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptSchemeValue](),
			},
			expectedPaths: []string{"metadata.taxonomy[0].taxonomy_concept.id"},
		},
		"unknown required": {
			value: ContentTypeMetadataTaxonomyItemValue{
				TaxonomyConcept: NewTypedObject(ContentTypeMetadataTaxonomyItemConceptValue{
					ID:       types.StringValue("concept"),
					Required: types.BoolUnknown(),
				}),
				TaxonomyConceptScheme: NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptSchemeValue](),
			},
			expectedPaths: []string{"metadata.taxonomy[0].taxonomy_concept.required"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := test.value.ToContentTypeMetadataTaxonomyItem(valuePath)

			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.expectedPaths, attributeDiagnosticPaths(t, diags))
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
	assert.Equal(t, []string{"metadata.taxonomy[1].taxonomy_concept.id"}, attributeDiagnosticPaths(t, diags))
}
