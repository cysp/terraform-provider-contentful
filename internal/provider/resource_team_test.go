package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccTeamResource(t *testing.T) {
	t.Parallel()

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

func TestAccTeamResourceImport(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	server.SetTeam("2zuSjSO4A0e6GKBrhJRe2m", "team-id", cm.TeamData{
		Name:        "Test Team",
		Description: cm.NewNilString(""),
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:             testAccTeamResourceImportConfig(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config:        testAccTeamResourceImportConfig(),
				ResourceName:  "contentful_team.test",
				ImportState:   true,
				ImportStateId: "2zuSjSO4A0e6GKBrhJRe2m/team-id",
				Check:         testAccTeamResourceImportCheck(),
			},
		},
	})
}

func testAccTeamResourceImportCheck() resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("contentful_team.test", "organization_id", "2zuSjSO4A0e6GKBrhJRe2m"),
		resource.TestCheckResourceAttr("contentful_team.test", "team_id", "team-id"),
		resource.TestCheckResourceAttr("contentful_team.test", "name", "Test Team"),
		resource.TestCheckResourceAttr("contentful_team.test", "description", ""),
	)
}

func testAccTeamResourceImportConfig() string {
	return `
resource "contentful_team" "test" {
  organization_id = "2zuSjSO4A0e6GKBrhJRe2m"

  name        = "Test Team"
  description = ""
}
`
}
