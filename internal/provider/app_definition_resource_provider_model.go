package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AppDefinitionResourceProviderModel struct {
	ID                 types.String `tfsdk:"id"`
	OrganizationID     types.String `tfsdk:"organization_id"`
	AppDefinitionID    types.String `tfsdk:"app_definition_id"`
	ResourceProviderID types.String `tfsdk:"resource_provider_id"`
	FunctionID         types.String `tfsdk:"function_id"`
}
