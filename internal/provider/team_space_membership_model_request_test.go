package provider_test

import (
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamSpaceMembershipModelToRequestRejectsNullAndUnknownRoles(t *testing.T) {
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
	require.True(t, diags.HasError())
	assert.Len(t, diags.Errors(), 2)

	assert.False(t, request.Admin)
	assert.Nil(t, request.Roles)
	assert.Equal(t, []string{"roles[1]", "roles[2]"}, diagnosticPaths(t, diags))
}
