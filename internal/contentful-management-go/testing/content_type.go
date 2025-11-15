package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

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

	contentType.Metadata = contentTypeFields.Metadata
}

func publishContentType(contentType *cm.ContentType) {
	contentType.Sys.PublishedVersion.SetTo(contentType.Sys.Version)

	contentType.Sys.Version++
}
