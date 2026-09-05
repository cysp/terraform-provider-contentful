package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccTeamSpaceMembershipResourceLifecycle(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	server.RegisterSpaceEnvironment("space-id", "master")

	identity := statecheck.CompareValue(compare.ValuesSame())

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: teamSpaceMembershipConfigVariables("space-id", "team-id", true),
				ConfigStateChecks: []statecheck.StateCheck{
					identity.AddStateValue("contentful_team_space_membership.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("contentful_team_space_membership.test", tfjsonpath.New("admin"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("contentful_team_space_membership.test", tfjsonpath.New("team_id"), knownvalue.StringExact("team-id")),
				},
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: teamSpaceMembershipConfigVariables("space-id", "team-id", false),
				ConfigStateChecks: []statecheck.StateCheck{
					identity.AddStateValue("contentful_team_space_membership.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("contentful_team_space_membership.test", tfjsonpath.New("admin"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("contentful_team_space_membership.test", tfjsonpath.New("team_id"), knownvalue.StringExact("team-id")),
				},
			},
		},
	})
}

func TestAccTeamSpaceMembershipResourceImport(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	server.RegisterSpaceEnvironment("space-id", "master")

	server.SetTeamSpaceMembership("space-id", "team-space-membership-id", "team-id", cm.TeamSpaceMembershipData{
		Admin: true,
		Roles: []cm.RoleLink{},
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ResourceName:    "contentful_team_space_membership.test",
				ImportState:     true,
				ImportStateId:   "space-id/team-space-membership-id",
				ConfigVariables: teamSpaceMembershipConfigVariables("space-id", "team-id", true),
				ImportStateCheck: testAccImportAttributes(map[string]string{
					"id":                       "space-id/team-space-membership-id",
					"space_id":                 "space-id",
					"team_space_membership_id": "team-space-membership-id",
					"team_id":                  "team-id",
					"admin":                    "true",
					"roles.#":                  "0",
				}),
			},
		},
	})
}

func teamSpaceMembershipConfigVariables(spaceID, teamID string, admin bool) config.Variables {
	return config.Variables{
		"space_id": config.StringVariable(spaceID),
		"team_id":  config.StringVariable(teamID),
		"admin":    config.BoolVariable(admin),
	}
}
