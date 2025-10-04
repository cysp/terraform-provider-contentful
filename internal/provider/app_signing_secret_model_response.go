package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewAppSigningSecretResourceModelFromResponse(_ context.Context, res cm.AppSigningSecret) (AppSigningSecretModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	organizationID := res.Sys.Organization.Sys.ID
	appDefinitionID := res.Sys.AppDefinition.Sys.ID

	model := AppSigningSecretModel{
		IDIdentityModel: IDIdentityModel{
			ID: types.StringValue(organizationID + "/" + appDefinitionID),
		},
		AppSigningSecretIdentityModel: AppSigningSecretIdentityModel{
			OrganizationID:  types.StringValue(organizationID),
			AppDefinitionID: types.StringValue(appDefinitionID),
		},
	}

	return model, diags
}
