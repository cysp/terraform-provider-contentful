package cmtesting

import (
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

const appSigningSecretMockUserID = "mock-user"

func NewAppSigningSecretFromRequest(organizationID, appDefinitionID string, request cm.AppSigningSecretRequestData) cm.AppSigningSecret {
	resourceProvider := cm.AppSigningSecret{
		Sys: cm.NewAppSigningSecretSys(organizationID, appDefinitionID, appSigningSecretMockUserID),
	}

	UpdateAppSigningSecretFromRequest(&resourceProvider, organizationID, appDefinitionID, request)

	return resourceProvider
}

func UpdateAppSigningSecretFromRequest(appSigningSecret *cm.AppSigningSecret, organizationID, appDefinitionID string, request cm.AppSigningSecretRequestData) {
	appSigningSecret.Sys.Organization.Sys.ID = organizationID
	appSigningSecret.Sys.AppDefinition.Sys.ID = appDefinitionID
	appSigningSecret.Sys.UpdatedAt = time.Now().UTC()
	appSigningSecret.Sys.UpdatedBy = cm.NewUserLink(appSigningSecretMockUserID)

	appSigningSecret.RedactedValue = request.Value[len(request.Value)-4:]
}
