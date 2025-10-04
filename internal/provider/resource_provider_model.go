package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ResourceProviderIdentityModel struct {
	OrganizationID  types.String `tfsdk:"organization_id"`
	AppDefinitionID types.String `tfsdk:"app_definition_id"`
}

type ResourceProviderModel struct {
	IDIdentityModel
	ResourceProviderIdentityModel

	ResourceProviderID types.String `tfsdk:"resource_provider_id"`
	FunctionID         types.String `tfsdk:"function_id"`
}
