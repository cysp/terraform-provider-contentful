package testing

import (
	"encoding/base64"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewAppSigningSecretFromRequest(organizationID, appDefinitionID string, request cm.AppSigningSecretRequestFields) cm.AppSigningSecret {
	resourceProvider := cm.AppSigningSecret{
		Sys: NewAppSigningSecretSys(organizationID, appDefinitionID),
	}

	UpdateAppSigningSecretFromRequest(&resourceProvider, organizationID, appDefinitionID, request)

	return resourceProvider
}

func NewAppSigningSecretSys(organizationID, appDefinitionID string) cm.AppSigningSecretSys {
	return cm.AppSigningSecretSys{
		Type: cm.AppSigningSecretSysTypeAppSigningSecret,
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
	}
}

func UpdateAppSigningSecretFromRequest(appSigningSecret *cm.AppSigningSecret, organizationID, appDefinitionID string, request cm.AppSigningSecretRequestFields) {
	appSigningSecret.Sys.Organization.Sys.ID = organizationID
	appSigningSecret.Sys.AppDefinition.Sys.ID = appDefinitionID

	appSigningSecret.RedactedValue = base64.StdEncoding.EncodeToString([]byte(request.Value))
}
