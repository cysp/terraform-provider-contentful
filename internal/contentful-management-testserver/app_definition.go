package contentfulmanagementtestserver

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewAppDefinitionFromFields(organizationID, appDefinitionID string, fields cm.AppDefinitionFields) cm.AppDefinition {
	appDefinition := cm.AppDefinition{
		Sys: NewAppDefinitionSys(organizationID, appDefinitionID),
	}

	UpdateAppDefinitionFromFields(&appDefinition, organizationID, appDefinitionID, fields)

	return appDefinition
}

func NewAppDefinitionSys(organizationID, appDefinitionID string) cm.AppDefinitionSys {
	return cm.AppDefinitionSys{
		Type: cm.AppDefinitionSysTypeAppDefinition,
		Organization: cm.OrganizationLink{
			Sys: cm.OrganizationLinkSys{
				Type:     cm.OrganizationLinkSysTypeLink,
				LinkType: cm.OrganizationLinkSysLinkTypeOrganization,
				ID:       organizationID,
			},
		},
		ID: appDefinitionID,
	}
}

func UpdateAppDefinitionFromFields(appDefinition *cm.AppDefinition, organizationID, appDefinitionID string, fields cm.AppDefinitionFields) {
	appDefinition.Sys.ID = appDefinitionID
	appDefinition.Sys.Organization.Sys.ID = organizationID

	appDefinition.Name = fields.Name

	appDefinition.Src = fields.Src
	appDefinition.Bundle = fields.Bundle

	appDefinition.Locations = convertSlice(fields.Locations, func(item cm.AppDefinitionFieldsLocationsItem) cm.AppDefinitionLocationsItem {
		locationsItem := cm.AppDefinitionLocationsItem{
			Location: item.Location,
			FieldTypes: convertSlice(item.FieldTypes, func(fieldType cm.AppDefinitionFieldsLocationsItemFieldTypesItem) cm.AppDefinitionLocationsItemFieldTypesItem {
				fieldTypesItem := cm.AppDefinitionLocationsItemFieldTypesItem{
					Type:     fieldType.Type,
					LinkType: fieldType.LinkType,
				}
				convertOptNil(&fieldTypesItem.Items, &fieldType.Items, func(items cm.AppDefinitionFieldsLocationsItemFieldTypesItemItems) cm.AppDefinitionLocationsItemFieldTypesItemItems {
					return cm.AppDefinitionLocationsItemFieldTypesItemItems(items)
				})

				return fieldTypesItem
			}),
		}

		convertOptNil(&locationsItem.NavigationItem, &item.NavigationItem, func(navigationItem cm.AppDefinitionFieldsLocationsItemNavigationItem) cm.AppDefinitionLocationsItemNavigationItem {
			return cm.AppDefinitionLocationsItemNavigationItem(navigationItem)
		})

		return locationsItem
	})

	appDefinition.Parameters = fields.Parameters
}
