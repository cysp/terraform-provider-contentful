package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewExtensionFromFields(spaceID, environmentID, extensionID string, fields cm.ExtensionData) cm.Extension {
	extension := cm.Extension{
		Sys: NewExtensionSys(spaceID, environmentID, extensionID),
	}

	UpdateExtensionFromFields(&extension, fields)

	return extension
}

func NewExtensionSys(spaceID, environmentID, extensionID string) cm.ExtensionSys {
	return cm.ExtensionSys{
		Type: cm.ExtensionSysTypeExtension,
		Space: cm.SpaceLink{
			Sys: cm.SpaceLinkSys{
				Type:     cm.SpaceLinkSysTypeLink,
				LinkType: cm.SpaceLinkSysLinkTypeSpace,
				ID:       spaceID,
			},
		},
		Environment: cm.EnvironmentLink{
			Sys: cm.EnvironmentLinkSys{
				Type:     cm.EnvironmentLinkSysTypeLink,
				LinkType: cm.EnvironmentLinkSysLinkTypeEnvironment,
				ID:       environmentID,
			},
		},
		ID: extensionID,
	}
}

func UpdateExtensionFromFields(extension *cm.Extension, fields cm.ExtensionData) {
	extensionExtension := cm.ExtensionExtension{
		Name:       fields.Extension.Name,
		Src:        fields.Extension.Src,
		Srcdoc:     fields.Extension.Srcdoc,
		Sidebar:    fields.Extension.Sidebar,
		Parameters: fields.Extension.Parameters,
	}

	if fields.Extension.FieldTypes != nil {
		fieldTypes := make([]cm.ExtensionExtensionFieldTypesItem, 0, len(fields.Extension.FieldTypes))

		for _, fieldType := range fields.Extension.FieldTypes {
			fieldTypesItem := cm.ExtensionExtensionFieldTypesItem{
				Type:     fieldType.Type,
				LinkType: cm.NewOptPointerString(fieldType.LinkType.ValueStringPointer()),
			}

			if fieldTypeItems, ok := fieldType.Items.Get(); ok {
				fieldTypesItem.Items = cm.NewOptExtensionExtensionFieldTypesItemItems(cm.ExtensionExtensionFieldTypesItemItems{
					Type:     fieldTypeItems.Type,
					LinkType: cm.NewOptPointerString(fieldTypeItems.LinkType.ValueStringPointer()),
				})
			}

			fieldTypes = append(fieldTypes, fieldTypesItem)
		}

		extensionExtension.FieldTypes = fieldTypes
	}

	extension.Extension = extensionExtension
	extension.Parameters = fields.Parameters
}
