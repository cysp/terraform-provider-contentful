package contentfulmanagementtestserver

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewAppDefinitionResourceProviderFromRequest(organizationID, appDefinitionID string, resourceProviderRequest cm.ResourceProviderRequest) cm.ResourceProvider {
	resourceProvider := cm.ResourceProvider{
		Sys: NewResourceProviderSys(organizationID, appDefinitionID, resourceProviderRequest.Sys.ID),
	}

	UpdateAppDefinitionResourceProviderFromRequest(&resourceProvider, organizationID, appDefinitionID, resourceProviderRequest)

	return resourceProvider
}

func NewResourceProviderSys(organizationID, appDefinitionID, resourceProviderID string) cm.ResourceProviderSys {
	return cm.ResourceProviderSys{
		Type: cm.ResourceProviderSysTypeResourceProvider,
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
		ID: resourceProviderID,
	}
}

func UpdateAppDefinitionResourceProviderFromRequest(resourceProvider *cm.ResourceProvider, organizationID, appDefinitionID string, resourceProviderRequest cm.ResourceProviderRequest) {
	resourceProvider.Sys.ID = resourceProviderRequest.Sys.ID
	resourceProvider.Sys.Organization.Sys.ID = organizationID
	resourceProvider.Sys.AppDefinition.Sys.ID = appDefinitionID

	resourceProvider.Type = cm.ResourceProviderType(resourceProviderRequest.Type)

	resourceProvider.Function.Sys.Type = resourceProviderRequest.Function.Sys.Type
	resourceProvider.Function.Sys.LinkType = resourceProviderRequest.Function.Sys.LinkType
	resourceProvider.Function.Sys.ID = resourceProviderRequest.Function.Sys.ID
}
