package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AppSigningSecretIdentityModel struct {
	OrganizationID  types.String `tfsdk:"organization_id"`
	AppDefinitionID types.String `tfsdk:"app_definition_id"`
}

type AppSigningSecretModel struct {
	IDIdentityModel
	AppSigningSecretIdentityModel

	Value   types.String `tfsdk:"value"`
	ValueWO types.String `tfsdk:"value_wo"`

	CreatedAt timetypes.RFC3339 `tfsdk:"created_at"`
	UpdatedAt timetypes.RFC3339 `tfsdk:"updated_at"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
