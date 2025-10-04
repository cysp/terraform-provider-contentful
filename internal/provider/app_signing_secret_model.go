package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AppSigningSecretIdentityModel struct {
	OrganizationID  types.String `tfsdk:"organization_id"`
	AppDefinitionID types.String `tfsdk:"app_definition_id"`
}

type AppSigningSecretModel struct {
	IDIdentityModel
	AppSigningSecretIdentityModel

	Value types.String `tfsdk:"value"`
}
