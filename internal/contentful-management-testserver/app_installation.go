package contentfulmanagementtestserver

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewAppInstallationFromFields(spaceID, environmentID, appDefinitionID string, appInstallationFields cm.AppInstallationFields) cm.AppInstallation {
	appInstallation := cm.AppInstallation{
		Sys: NewAppInstallationSys(spaceID, environmentID, appDefinitionID),
	}

	UpdateAppInstallationFromFields(&appInstallation, appInstallationFields)

	return appInstallation
}

func NewAppInstallationSys(spaceID, environmentID, appDefinitionID string) cm.AppInstallationSys {
	return cm.AppInstallationSys{
		Type: cm.AppInstallationSysTypeAppInstallation,
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
		AppDefinition: cm.AppDefinitionLink{
			Sys: cm.AppDefinitionLinkSys{
				Type:     cm.AppDefinitionLinkSysTypeLink,
				LinkType: cm.AppDefinitionLinkSysLinkTypeAppDefinition,
				ID:       appDefinitionID,
			},
		},
	}
}

func UpdateAppInstallationFromFields(appInstallation *cm.AppInstallation, appInstallationFields cm.AppInstallationFields) {
	appInstallation.Parameters = appInstallationFields.Parameters
}
