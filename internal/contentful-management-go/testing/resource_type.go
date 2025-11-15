package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewResourceTypeFromRequest(organizationID, appDefinitionID, resourceProviderID, resourceTypeID string, fields cm.ResourceTypeData) cm.ResourceType {
	resourceProvider := cm.ResourceType{
		Sys: cm.NewResourceTypeSys(organizationID, appDefinitionID, resourceProviderID, resourceTypeID),
	}

	UpdateResourceTypeFromFields(&resourceProvider, organizationID, appDefinitionID, resourceProviderID, resourceTypeID, fields)

	return resourceProvider
}

func UpdateResourceTypeFromFields(resourceType *cm.ResourceType, organizationID, appDefinitionID, resourceProviderID, resourceTypeID string, fields cm.ResourceTypeData) {
	resourceType.Sys.ID = resourceTypeID
	resourceType.Sys.Organization.Sys.ID = organizationID
	resourceType.Sys.AppDefinition.Sys.ID = appDefinitionID
	resourceType.Sys.ResourceProvider.Sys.ID = resourceProviderID

	resourceType.Name = fields.Name
	resourceType.DefaultFieldMapping = fields.DefaultFieldMapping
}
