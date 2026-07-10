//nolint:testpackage
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/stretchr/testify/assert"
)

func TestReconcileContentTypeMetadataPlan(t *testing.T) {
	t.Parallel()

	taxonomyNull := NewTypedListNull[TypedObject[ContentTypeMetadataTaxonomyItemValue]]()
	taxonomyEmpty := NewTypedList([]TypedObject[ContentTypeMetadataTaxonomyItemValue]{})
	taxonomyPopulated := NewTypedList([]TypedObject[ContentTypeMetadataTaxonomyItemValue]{
		NewTypedObject(ContentTypeMetadataTaxonomyItemValue{}),
	})
	annotations := NewNormalizedJSONTypesNormalizedValue([]byte(`{"ContentType":[]}`))

	metadata := func(annotations jsontypes.Normalized, taxonomy TypedList[TypedObject[ContentTypeMetadataTaxonomyItemValue]]) TypedObject[ContentTypeMetadataValue] {
		return NewTypedObject(ContentTypeMetadataValue{
			Annotations: annotations,
			Taxonomy:    taxonomy,
		})
	}

	tests := map[string]struct {
		configMetadata   TypedObject[ContentTypeMetadataValue]
		stateMetadata    TypedObject[ContentTypeMetadataValue]
		expectedMetadata TypedObject[ContentTypeMetadataValue]
		expectedModified bool
	}{
		"unknown configuration": {
			configMetadata:   NewTypedObjectUnknown[ContentTypeMetadataValue](),
			stateMetadata:    metadata(annotations, taxonomyPopulated),
			expectedModified: false,
		},
		"absent state": {
			configMetadata:   NewTypedObjectNull[ContentTypeMetadataValue](),
			stateMetadata:    NewTypedObjectNull[ContentTypeMetadataValue](),
			expectedModified: false,
		},
		"omitted metadata clears annotations without taxonomy": {
			configMetadata:   NewTypedObjectNull[ContentTypeMetadataValue](),
			stateMetadata:    metadata(annotations, taxonomyNull),
			expectedMetadata: NewTypedObjectUnknown[ContentTypeMetadataValue](),
			expectedModified: true,
		},
		"omitted metadata preserves taxonomy": {
			configMetadata: NewTypedObjectNull[ContentTypeMetadataValue](),
			stateMetadata:  metadata(annotations, taxonomyPopulated),
			expectedMetadata: metadata(
				jsontypes.NewNormalizedNull(),
				taxonomyPopulated,
			),
			expectedModified: true,
		},
		"omitted metadata preserves empty taxonomy": {
			configMetadata: NewTypedObjectNull[ContentTypeMetadataValue](),
			stateMetadata:  metadata(annotations, taxonomyEmpty),
			expectedMetadata: metadata(
				jsontypes.NewNormalizedNull(),
				taxonomyEmpty,
			),
			expectedModified: true,
		},
		"omitted taxonomy preserves prior taxonomy": {
			configMetadata: metadata(annotations, taxonomyNull),
			stateMetadata:  metadata(jsontypes.NewNormalizedNull(), taxonomyPopulated),
			expectedMetadata: metadata(
				annotations,
				taxonomyPopulated,
			),
			expectedModified: true,
		},
		"explicit empty taxonomy remains configured": {
			configMetadata:   metadata(jsontypes.NewNormalizedNull(), taxonomyEmpty),
			stateMetadata:    metadata(annotations, taxonomyPopulated),
			expectedModified: false,
		},
		"explicit populated taxonomy remains configured": {
			configMetadata:   metadata(jsontypes.NewNormalizedNull(), taxonomyPopulated),
			stateMetadata:    metadata(annotations, taxonomyEmpty),
			expectedModified: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actualMetadata, actualModified := reconcileContentTypeMetadataPlan(test.configMetadata, test.stateMetadata)

			assert.Equal(t, test.expectedMetadata, actualMetadata)
			assert.Equal(t, test.expectedModified, actualModified)
		})
	}
}
