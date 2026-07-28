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

func TestTeamSpaceMembershipModelToRequestRoles(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		roles               []types.String
		expectedRoles       []cm.RoleLink
		expectedDiagnostics []string
	}{
		"known": {
			roles: []types.String{
				types.StringValue("role-a"),
				types.StringValue("role-b"),
			},
			expectedRoles: []cm.RoleLink{
				{Sys: cm.RoleLinkSys{Type: cm.RoleLinkSysTypeLink, LinkType: cm.RoleLinkSysLinkTypeRole, ID: "role-a"}},
				{Sys: cm.RoleLinkSys{Type: cm.RoleLinkSysTypeLink, LinkType: cm.RoleLinkSysLinkTypeRole, ID: "role-b"}},
			},
		},
		"known empty": {
			roles:         []types.String{},
			expectedRoles: []cm.RoleLink{},
		},
		"invalid children": {
			roles: []types.String{
				types.StringValue("role-a"),
				types.StringNull(),
				types.StringUnknown(),
				types.StringValue("role-b"),
			},
			expectedDiagnostics: []string{"roles[1]", "roles[2]"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := provider.TeamSpaceMembershipModel{
				Admin: types.BoolValue(true),
				Roles: test.roles,
			}

			request, diags := model.ToTeamSpaceMembershipData(t.Context(), path.Empty())

			if len(test.expectedDiagnostics) == 0 {
				require.False(t, diags.HasError(), diags.Errors())
				assert.True(t, request.Admin)
				require.NotNil(t, request.Roles)
				assert.Equal(t, test.expectedRoles, request.Roles)

				return
			}

			assert.Equal(t, cm.TeamSpaceMembershipData{}, request)
			require.True(t, diags.HasError())
			assert.Equal(t, test.expectedDiagnostics, attributeDiagnosticPaths(t, diags))
		})
	}
}
