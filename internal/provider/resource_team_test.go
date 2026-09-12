package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccTeamResourceLifecycle(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	identity := statecheck.CompareValue(compare.ValuesSame())

	testAccMockedResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id": config.StringVariable("2zuSjSO4A0e6GKBrhJRe2m"),
					"team_name":       config.StringVariable("Test Team"),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.AddStateValue("contentful_team.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("contentful_team.test", tfjsonpath.New("description"), knownvalue.StringExact("")),
				},
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id": config.StringVariable("2zuSjSO4A0e6GKBrhJRe2m"),
					"team_name":       config.StringVariable("Test Team Updated"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_team.test", plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.AddStateValue("contentful_team.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("contentful_team.test", tfjsonpath.New("description"), knownvalue.StringExact("")),
				},
			},
		},
	})
}

func TestAccTeamResourceImport(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	server.SetTeam("2zuSjSO4A0e6GKBrhJRe2m", "team-id", cm.TeamData{
		Name:        "Test Team",
		Description: cm.NewNilString(""),
	})

	testAccMockedResource(t, server, resource.TestCase{
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
				ImportStateCheck: testAccImportAttributes(map[string]string{
					"organization_id": "2zuSjSO4A0e6GKBrhJRe2m",
					"team_id":         "team-id",
					"name":            "Test Team",
					"description":     "",
				}),
			},
		},
	})
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

func TestAccTeamResourceImportUpdateUsesDefaultDescription(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	server.SetTeam("2zuSjSO4A0e6GKBrhJRe2m", "team-id", cm.TeamData{
		Name:        "Test Team",
		Description: cm.NewNilString("Existing description"),
	})

	testAccMockedResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `
resource "contentful_team" "test" {
  organization_id = "2zuSjSO4A0e6GKBrhJRe2m"

  name = "Test Team"
}
`,
				ResourceName:       "contentful_team.test",
				ImportState:        true,
				ImportStateId:      "2zuSjSO4A0e6GKBrhJRe2m/team-id",
				ImportStatePersist: true,
			},
			{
				Config: `
resource "contentful_team" "test" {
  organization_id = "2zuSjSO4A0e6GKBrhJRe2m"

  name = "Test Team Updated"
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_team.test", tfjsonpath.New("description"), knownvalue.StringExact("")),
				},
			},
		},
	})
}

func TestAccTeamResourceImportNullDescriptionUsesDefault(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	server.SetTeam("2zuSjSO4A0e6GKBrhJRe2m", "team-id", cm.TeamData{
		Name:        "Test Team",
		Description: cm.NewNilStringNull(),
	})

	testAccMockedResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `
resource "contentful_team" "test" {
  organization_id = "2zuSjSO4A0e6GKBrhJRe2m"

  name = "Test Team"
}
`,
				ResourceName:       "contentful_team.test",
				ImportState:        true,
				ImportStateId:      "2zuSjSO4A0e6GKBrhJRe2m/team-id",
				ImportStatePersist: true,
			},
			{
				Config: `
resource "contentful_team" "test" {
  organization_id = "2zuSjSO4A0e6GKBrhJRe2m"

  name = "Test Team"
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_team.test", tfjsonpath.New("description"), knownvalue.StringExact("")),
				},
			},
		},
	})
}
