package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//nolint:paralleltest
func TestAccTeamResource(t *testing.T) {
	server, _ := cmt.NewContentfulManagementServer()

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id": config.StringVariable("2zuSjSO4A0e6GKBrhJRe2m"),
					"team_name":       config.StringVariable("Test Team"),
				},
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id": config.StringVariable("2zuSjSO4A0e6GKBrhJRe2m"),
					"team_name":       config.StringVariable("Test Team Updated"),
				},
			},
		},
	})
}

//nolint:paralleltest
func TestAccTeamResourceImport(t *testing.T) {
	server, _ := cmt.NewContentfulManagementServer()

	server.SetTeam("2zuSjSO4A0e6GKBrhJRe2m", "team-id", cm.TeamData{
		Name: "Test Team",
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id": config.StringVariable("2zuSjSO4A0e6GKBrhJRe2m"),
					"team_name":       config.StringVariable("Test Team"),
				},
			},
		},
	})
}
