package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//nolint:paralleltest
func TestAccTeamSpaceMembershipResource(t *testing.T) {
	server, _ := cmt.NewContentfulManagementServer()

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable("space-id"),
					"team_id":  config.StringVariable("team-id"),
					"admin":    config.BoolVariable(true),
				},
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable("space-id"),
					"team_id":  config.StringVariable("team-id"),
					"admin":    config.BoolVariable(false),
				},
			},
		},
	})
}

//nolint:paralleltest
func TestAccTeamSpaceMembershipResourceImport(t *testing.T) {
	server, _ := cmt.NewContentfulManagementServer()

	server.SetTeamSpaceMembership("space-id", "team-space-membership-id", "team-id", cm.TeamSpaceMembershipData{
		Admin: true,
		Roles: []cm.RoleLink{},
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable("space-id"),
					"team_id":  config.StringVariable("team-id"),
					"admin":    config.BoolVariable(true),
				},
			},
		},
	})
}
