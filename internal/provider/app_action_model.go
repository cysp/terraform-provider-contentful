package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AppActionFields struct {
	AppActionID      types.String         `tfsdk:"app_action_id"`
	Name             types.String         `tfsdk:"name"`
	Category         types.String         `tfsdk:"category"`
	Type             types.String         `tfsdk:"type"`
	Description      types.String         `tfsdk:"description"`
	URL              types.String         `tfsdk:"url"`
	FunctionID       types.String         `tfsdk:"function_id"`
	Parameters       jsontypes.Normalized `tfsdk:"parameters"`
	ParametersSchema jsontypes.Normalized `tfsdk:"parameters_schema"`
	ResultSchema     jsontypes.Normalized `tfsdk:"result_schema"`
}

type AppActionBaseModel struct {
	IDIdentityModel
	AppActionFields

	OrganizationID  types.String `tfsdk:"organization_id"`
	AppDefinitionID types.String `tfsdk:"app_definition_id"`
}

type AppActionModel struct {
	AppActionBaseModel

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
