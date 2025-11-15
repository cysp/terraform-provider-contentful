package testing

import (
	"encoding/base64"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewAppSigningSecretFromRequest(organizationID, appDefinitionID string, request cm.AppSigningSecretRequestData) cm.AppSigningSecret {
	resourceProvider := cm.AppSigningSecret{
		Sys: cm.NewAppSigningSecretSys(organizationID, appDefinitionID),
	}

	UpdateAppSigningSecretFromRequest(&resourceProvider, organizationID, appDefinitionID, request)

	return resourceProvider
}

func UpdateAppSigningSecretFromRequest(appSigningSecret *cm.AppSigningSecret, organizationID, appDefinitionID string, request cm.AppSigningSecretRequestData) {
	appSigningSecret.Sys.Organization.Sys.ID = organizationID
	appSigningSecret.Sys.AppDefinition.Sys.ID = appDefinitionID

	appSigningSecret.RedactedValue = base64.StdEncoding.EncodeToString([]byte(request.Value))
}
