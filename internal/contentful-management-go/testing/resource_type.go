package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewResourceTypeFromRequest(organizationID, appDefinitionID, resourceProviderID, resourceTypeID string, fields cm.ResourceTypeData) cm.ResourceType {
	resourceProvider := cm.ResourceType{
		Sys: NewResourceTypeSys(organizationID, appDefinitionID, resourceProviderID, resourceTypeID),
	}

	UpdateResourceTypeFromFields(&resourceProvider, organizationID, appDefinitionID, resourceProviderID, resourceTypeID, fields)

	return resourceProvider
}

func NewResourceTypeSys(organizationID, appDefinitionID, resourceProviderID, resourceTypeID string) cm.ResourceTypeSys {
	return cm.ResourceTypeSys{
		Type: cm.ResourceTypeSysTypeResourceType,
		Organization: cm.OrganizationLink{
			Sys: cm.OrganizationLinkSys{
				Type:     cm.OrganizationLinkSysTypeLink,
				LinkType: cm.OrganizationLinkSysLinkTypeOrganization,
				ID:       organizationID,
			},
		},
		AppDefinition: cm.AppDefinitionLink{
			Sys: cm.AppDefinitionLinkSys{
				Type:     cm.AppDefinitionLinkSysTypeLink,
				LinkType: cm.AppDefinitionLinkSysLinkTypeAppDefinition,
				ID:       appDefinitionID,
			},
		},
		ResourceProvider: cm.ResourceProviderLink{
			Sys: cm.ResourceProviderLinkSys{
				Type:     cm.ResourceProviderLinkSysTypeLink,
				LinkType: cm.ResourceProviderLinkSysLinkTypeResourceProvider,
				ID:       resourceProviderID,
			},
		},
		ID: resourceTypeID,
	}
}

func UpdateResourceTypeFromFields(resourceType *cm.ResourceType, organizationID, appDefinitionID, resourceProviderID, resourceTypeID string, fields cm.ResourceTypeData) {
	resourceType.Sys.ID = resourceTypeID
	resourceType.Sys.Organization.Sys.ID = organizationID
	resourceType.Sys.AppDefinition.Sys.ID = appDefinitionID
	resourceType.Sys.ResourceProvider.Sys.ID = resourceProviderID

	resourceType.Name = fields.Name
	resourceType.DefaultFieldMapping = fields.DefaultFieldMapping
}
