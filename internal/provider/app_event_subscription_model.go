package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AppEventSubscriptionModel struct {
	IDIdentityModel

	OrganizationID           types.String   `tfsdk:"organization_id"`
	AppDefinitionID          types.String   `tfsdk:"app_definition_id"`
	Topics                   types.Set      `tfsdk:"topics"`
	TargetURL                types.String   `tfsdk:"target_url"`
	FilterFunctionID         types.String   `tfsdk:"filter_function_id"`
	TransformationFunctionID types.String   `tfsdk:"transformation_function_id"`
	HandlerFunctionID        types.String   `tfsdk:"handler_function_id"`
	Timeouts                 timeouts.Value `tfsdk:"timeouts"`
}
