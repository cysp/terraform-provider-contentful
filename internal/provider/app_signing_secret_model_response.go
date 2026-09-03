package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewAppSigningSecretResourceModelFromResponse(res cm.AppSigningSecret) AppSigningSecretModel {
	organizationID := res.Sys.Organization.Sys.ID
	appDefinitionID := res.Sys.AppDefinition.Sys.ID

	model := AppSigningSecretModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(organizationID, appDefinitionID),
		AppSigningSecretIdentityModel: AppSigningSecretIdentityModel{
			OrganizationID:  types.StringValue(organizationID),
			AppDefinitionID: types.StringValue(appDefinitionID),
		},
	}

	return model
}
