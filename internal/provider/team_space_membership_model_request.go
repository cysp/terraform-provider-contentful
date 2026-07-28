package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *TeamSpaceMembershipModel) ToTeamSpaceMembershipData(_ context.Context, modelPath path.Path) (cm.TeamSpaceMembershipData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	roleIDs, roleIDDiags := knownStringListElements(modelPath.AtName("roles"), model.Roles)
	diags.Append(roleIDDiags...)

	if diags.HasError() {
		return cm.TeamSpaceMembershipData{}, diags
	}

	roles := make([]cm.RoleLink, 0, len(roleIDs))

	for _, roleID := range roleIDs {
		roles = append(roles, cm.RoleLink{
			Sys: cm.RoleLinkSys{
				Type:     cm.RoleLinkSysTypeLink,
				LinkType: cm.RoleLinkSysLinkTypeRole,
				ID:       roleID,
			},
		})
	}

	return cm.TeamSpaceMembershipData{
		Admin: model.Admin.ValueBool(),
		Roles: roles,
	}, diags
}
