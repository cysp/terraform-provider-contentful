package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamSpaceMembershipModelToRequestSkipsNullAndUnknownRoles(t *testing.T) {
	t.Parallel()

	model := provider.TeamSpaceMembershipModel{
		Admin: types.BoolValue(true),
		Roles: []types.String{
			types.StringValue("role-a"),
			types.StringNull(),
			types.StringUnknown(),
			types.StringValue("role-b"),
		},
	}

	request, diags := model.ToTeamSpaceMembershipData(t.Context(), path.Empty())
	require.False(t, diags.HasError(), diags.Errors())

	assert.True(t, request.Admin)
	assert.Equal(t, []cm.RoleLink{
		{Sys: cm.RoleLinkSys{Type: cm.RoleLinkSysTypeLink, LinkType: cm.RoleLinkSysLinkTypeRole, ID: "role-a"}},
		{Sys: cm.RoleLinkSys{Type: cm.RoleLinkSysTypeLink, LinkType: cm.RoleLinkSysLinkTypeRole, ID: "role-b"}},
	}, request.Roles)
}
