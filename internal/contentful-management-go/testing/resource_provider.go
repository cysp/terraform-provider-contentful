package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewResourceProviderFromRequest(organizationID, appDefinitionID string, resourceProviderRequest cm.ResourceProviderRequest) cm.ResourceProvider {
	resourceProvider := cm.ResourceProvider{
		Sys: cm.NewResourceProviderSys(organizationID, appDefinitionID, resourceProviderRequest.Sys.ID),
	}

	UpdateResourceProviderFromRequest(&resourceProvider, organizationID, appDefinitionID, resourceProviderRequest)

	return resourceProvider
}

func UpdateResourceProviderFromRequest(resourceProvider *cm.ResourceProvider, organizationID, appDefinitionID string, resourceProviderRequest cm.ResourceProviderRequest) {
	resourceProvider.Sys.ID = resourceProviderRequest.Sys.ID
	resourceProvider.Sys.Organization.Sys.ID = organizationID
	resourceProvider.Sys.AppDefinition.Sys.ID = appDefinitionID

	resourceProvider.Type = cm.ResourceProviderType(resourceProviderRequest.Type)

	resourceProvider.Function.Sys.Type = resourceProviderRequest.Function.Sys.Type
	resourceProvider.Function.Sys.LinkType = resourceProviderRequest.Function.Sys.LinkType
	resourceProvider.Function.Sys.ID = resourceProviderRequest.Function.Sys.ID
}
