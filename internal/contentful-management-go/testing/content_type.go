package cmtesting

import (
	"bytes"
	"encoding/json"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func contentTypeMetadataIsEmpty(metadata cm.ContentTypeMetadata) bool {
	annotations := bytes.TrimSpace(metadata.Annotations)
	if len(annotations) == 0 {
		return metadata.Taxonomy == nil
	}

	var annotationAssignments map[string]json.RawMessage

	err := json.Unmarshal(annotations, &annotationAssignments)
	if err != nil {
		return false
	}

	return len(annotationAssignments) == 0 && metadata.Taxonomy == nil
}

func NewContentTypeFromRequestFields(spaceID, environmentID, contentTypeID string, contentTypeFields cm.ContentTypeRequestData) cm.ContentType {
	contentType := cm.ContentType{
		Sys: cm.NewContentTypeSys(spaceID, environmentID, contentTypeID),
	}

	UpdateContentTypeFromRequestFields(&contentType, contentTypeFields)

	return contentType
}

func UpdateContentTypeFromRequestFields(contentType *cm.ContentType, contentTypeFields cm.ContentTypeRequestData) {
	contentType.Sys.Version++

	contentType.Name = contentTypeFields.Name

	contentType.Description = contentTypeFields.Description

	contentType.Fields = convertSlice(contentTypeFields.Fields, func(field cm.ContentTypeRequestDataFieldsItem) cm.ContentTypeFieldsItem {
		contentTypeFieldItems := cm.OptContentTypeFieldsItemItems{}
		convertOptNil(&contentTypeFieldItems, &field.Items, func(fieldItems cm.ContentTypeRequestDataFieldsItemItems) cm.ContentTypeFieldsItemItems {
			return cm.ContentTypeFieldsItemItems(fieldItems)
		})

		return cm.ContentTypeFieldsItem{
			ID:               field.ID,
			Name:             field.Name,
			Type:             field.Type,
			LinkType:         field.LinkType,
			Items:            contentTypeFieldItems,
			Localized:        field.Localized,
			Required:         field.Required,
			Validations:      field.Validations,
			Omitted:          field.Omitted,
			Disabled:         field.Disabled,
			DefaultValue:     field.DefaultValue,
			AllowedResources: field.AllowedResources,
		}
	})

	contentType.DisplayField = cm.NewNilString(contentTypeFields.DisplayField)

	updateContentTypeMetadata(contentType, contentTypeFields.Metadata)
}

func updateContentTypeMetadata(contentType *cm.ContentType, requestedMetadata cm.OptContentTypeMetadata) {
	existingMetadata, existingMetadataSet := contentType.Metadata.Get()
	metadata, metadataSet := requestedMetadata.Get()

	if !metadataSet {
		if !existingMetadataSet {
			return
		}

		existingMetadata.Annotations = nil
		if existingMetadata.Taxonomy == nil {
			contentType.Metadata.Reset()

			return
		}

		contentType.Metadata.SetTo(existingMetadata)

		return
	}

	if metadata.Taxonomy == nil && existingMetadataSet {
		metadata.Taxonomy = existingMetadata.Taxonomy
	}

	contentType.Metadata.SetTo(metadata)
}

func publishContentType(contentType *cm.ContentType, publishedAt time.Time) {
	contentType.Sys.PublishedVersion.SetTo(contentType.Sys.Version)
	contentType.Sys.PublishedAt.SetTo(publishedAt)

	contentType.Sys.Version++
}
