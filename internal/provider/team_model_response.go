package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewTeamResourceModelFromResponse(_ context.Context, response cm.Team) (TeamModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	organizationID := response.Sys.Organization.Sys.ID
	teamID := response.Sys.ID

	model := TeamModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(organizationID, teamID),
		TeamIdentityModel: TeamIdentityModel{
			OrganizationID: types.StringValue(organizationID),
			TeamID:         types.StringValue(teamID),
		},
	}

	model.Name = types.StringValue(response.Name)

	model.Description = types.StringPointerValue(response.Description.ValueStringPointer())

	return model, diags
}
