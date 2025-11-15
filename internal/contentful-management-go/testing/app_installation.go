package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewAppInstallationFromFields(spaceID, environmentID, appDefinitionID string, appInstallationFields cm.AppInstallationData) cm.AppInstallation {
	appInstallation := cm.AppInstallation{
		Sys: cm.NewAppInstallationSys(spaceID, environmentID, appDefinitionID),
	}

	UpdateAppInstallationFromFields(&appInstallation, appInstallationFields)

	return appInstallation
}

func UpdateAppInstallationFromFields(appInstallation *cm.AppInstallation, appInstallationFields cm.AppInstallationData) {
	appInstallation.Parameters = appInstallationFields.Parameters
}
