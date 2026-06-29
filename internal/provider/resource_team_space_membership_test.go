package provider_test

import (
	"errors"
	"fmt"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	errUnexpectedImportStateCount = errors.New("unexpected import state count")
	errUnexpectedImportedAttr     = errors.New("unexpected imported attribute")
)

func TestAccTeamSpaceMembershipResource(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.RegisterSpaceEnvironment("space-id", "master")

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: teamSpaceMembershipConfigVariables("space-id", "team-id", true),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: teamSpaceMembershipConfigVariables("space-id", "team-id", false),
			},
		},
	})
}

func TestAccTeamSpaceMembershipResourceImport(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.RegisterSpaceEnvironment("space-id", "master")

	server.SetTeamSpaceMembership("space-id", "team-space-membership-id", "team-id", cm.TeamSpaceMembershipData{
		Admin: true,
		Roles: []cm.RoleLink{},
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:  config.TestNameDirectory(),
				ResourceName:     "contentful_team_space_membership.test",
				ImportState:      true,
				ImportStateId:    "space-id/team-space-membership-id",
				ConfigVariables:  teamSpaceMembershipConfigVariables("space-id", "team-id", true),
				ImportStateCheck: testAccTeamSpaceMembershipImportCheck(),
			},
		},
	})
}

func testAccTeamSpaceMembershipImportCheck() resource.ImportStateCheckFunc {
	return func(states []*terraform.InstanceState) error {
		if len(states) != 1 {
			return fmt.Errorf("%w: expected 1 imported resource, got %d", errUnexpectedImportStateCount, len(states))
		}

		attributes := states[0].Attributes
		expectedAttributes := map[string]string{
			"id":                       "space-id/team-space-membership-id",
			"space_id":                 "space-id",
			"team_space_membership_id": "team-space-membership-id",
			"team_id":                  "team-id",
			"admin":                    "true",
			"roles.#":                  "0",
		}

		for name, expected := range expectedAttributes {
			if actual := attributes[name]; actual != expected {
				return fmt.Errorf("%w: expected %q to be %q, got %q", errUnexpectedImportedAttr, name, expected, actual)
			}
		}

		return nil
	}
}

func teamSpaceMembershipConfigVariables(spaceID, teamID string, admin bool) config.Variables {
	return config.Variables{
		"space_id": config.StringVariable(spaceID),
		"team_id":  config.StringVariable(teamID),
		"admin":    config.BoolVariable(admin),
	}
}
