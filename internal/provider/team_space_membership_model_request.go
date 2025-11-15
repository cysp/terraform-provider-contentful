package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *TeamSpaceMembershipModel) ToTeamSpaceMembershipData(_ context.Context, _ path.Path) (cm.TeamSpaceMembershipData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fields := cm.TeamSpaceMembershipData{
		Admin: model.Admin.ValueBool(),
	}

	if model.Roles != nil {
		roles := make([]cm.RoleLink, 0, len(model.Roles))
		for _, roleID := range model.Roles {
			if roleID.IsNull() || roleID.IsUnknown() {
				continue
			}

			roleLink := cm.RoleLink{
				Sys: cm.RoleLinkSys{
					Type:     cm.RoleLinkSysTypeLink,
					LinkType: cm.RoleLinkSysLinkTypeRole,
					ID:       roleID.ValueString(),
				},
			}

			roles = append(roles, roleLink)
		}

		fields.Roles = roles
	}

	return fields, diags
}
