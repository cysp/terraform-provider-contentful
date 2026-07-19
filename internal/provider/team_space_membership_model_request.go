package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *TeamSpaceMembershipModel) ToTeamSpaceMembershipData(_ context.Context, modelPath path.Path) (cm.TeamSpaceMembershipData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fields := cm.TeamSpaceMembershipData{
		Admin: model.Admin.ValueBool(),
	}

	if model.Roles != nil {
		roles := make([]cm.RoleLink, 0, len(model.Roles))
		for index, roleID := range model.Roles {
			roleIDString, roleIDDiags := KnownStringValue(roleID, modelPath.AtName("roles").AtListIndex(index))
			diags.Append(roleIDDiags...)

			if roleIDDiags.HasError() {
				continue
			}

			roleLink := cm.RoleLink{
				Sys: cm.RoleLinkSys{
					Type:     cm.RoleLinkSysTypeLink,
					LinkType: cm.RoleLinkSysLinkTypeRole,
					ID:       roleIDString,
				},
			}

			roles = append(roles, roleLink)
		}

		fields.Roles = roles
	}

	return fields, diags
}
