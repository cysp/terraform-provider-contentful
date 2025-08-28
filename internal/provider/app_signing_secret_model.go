package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AppSigningSecretModel struct {
	ID              types.String `tfsdk:"id"`
	OrganizationID  types.String `tfsdk:"organization_id"`
	AppDefinitionID types.String `tfsdk:"app_definition_id"`
	Value           types.String `tfsdk:"value"`
}
