package provider

import "github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"

func reconcileContentTypeMetadataPlan(
	configMetadata TypedObject[ContentTypeMetadataValue],
	stateMetadata TypedObject[ContentTypeMetadataValue],
) (TypedObject[ContentTypeMetadataValue], bool) {
	if configMetadata.IsUnknown() || stateMetadata.IsNull() || stateMetadata.IsUnknown() {
		return TypedObject[ContentTypeMetadataValue]{}, false
	}

	stateMetadataValue := stateMetadata.Value()

	if configMetadata.IsNull() {
		if stateMetadataValue.Taxonomy.IsNull() {
			return NewTypedObjectUnknown[ContentTypeMetadataValue](), true
		}

		return NewTypedObject(ContentTypeMetadataValue{
			Annotations: jsontypes.NewNormalizedNull(),
			Taxonomy:    stateMetadataValue.Taxonomy,
		}), true
	}

	configMetadataValue := configMetadata.Value()
	if configMetadataValue.Taxonomy.IsNull() && !stateMetadataValue.Taxonomy.IsNull() {
		configMetadataValue.Taxonomy = stateMetadataValue.Taxonomy

		return NewTypedObject(configMetadataValue), true
	}

	return TypedObject[ContentTypeMetadataValue]{}, false
}
