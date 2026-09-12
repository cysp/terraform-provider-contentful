package provider_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

var errUnexpectedTeamListRequestCount = errors.New("unexpected team-list request count")

// These acceptance tests deliberately use the mocked Contentful service. They
// validate Terraform behavior and the subset of the published HTTP contract
// consumed by the provider, not live authorization or visibility of
// SCIM-managed teams.

func TestAccTeamsDataSourceRead(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	organizationID := "2zuSjSO4A0e6GKBrhJRe2m"

	server.SetTeam(organizationID, "team-b", cm.TeamData{
		Name:        "Second Team",
		Description: cm.NewNilString("The second team."),
	})
	server.SetTeam(organizationID, "team-a", cm.TeamData{
		Name:        "First Team",
		Description: cm.NewNilString("The first team."),
	})

	testAccMockedResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id": config.StringVariable(organizationID),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.contentful_teams.test", tfjsonpath.New("id"), knownvalue.StringExact(organizationID)),
					statecheck.ExpectKnownValue("data.contentful_teams.test", tfjsonpath.New("organization_id"), knownvalue.StringExact(organizationID)),
					statecheck.ExpectKnownValue("data.contentful_teams.test", tfjsonpath.New("teams"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{"team_id": knownvalue.StringExact("team-a"), "name": knownvalue.StringExact("First Team"), "description": knownvalue.StringExact("The first team.")}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{"team_id": knownvalue.StringExact("team-b"), "name": knownvalue.StringExact("Second Team"), "description": knownvalue.StringExact("The second team.")}),
					})),
				},
			},
		},
	})
}

func TestAccTeamsDataSourceEmpty(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	organizationID := "2zuSjSO4A0e6GKBrhJRe2m"

	testAccMockedResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id": config.StringVariable(organizationID),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.contentful_teams.test", tfjsonpath.New("id"), knownvalue.StringExact(organizationID)),
					statecheck.ExpectKnownValue("data.contentful_teams.test", tfjsonpath.New("teams"), knownvalue.ListSizeExact(0)),
				},
			},
		},
	})
}

func TestAccTeamsDataSourcePagination(t *testing.T) {
	t.Parallel()

	organizationID := "2zuSjSO4A0e6GKBrhJRe2m"

	var requestCount atomic.Int64

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)

		if r.Method != http.MethodGet || r.URL.Path != "/organizations/"+organizationID+"/teams" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		if actual := r.Header.Get("Authorization"); actual != "Bearer "+cmt.ValidAccessToken {
			t.Errorf("unexpected authorization header: %q", actual)
		}

		if actual := r.URL.Query().Get("limit"); actual != "100" {
			t.Errorf("unexpected limit: %q", actual)
		}

		skip, err := strconv.Atoi(r.URL.Query().Get("skip"))
		if err != nil {
			t.Errorf("invalid skip: %v", err)
		}

		var items []map[string]any

		switch skip {
		case 0:
			items = make([]map[string]any, 0, 100)

			for i := 1; i <= 100; i++ {
				teamID := fmt.Sprintf("team-%03d", i)
				items = append(items, teamListItem(organizationID, teamID, teamID, "A test team."))
			}
		case 100:
			items = []map[string]any{
				teamListItem(organizationID, "team-000", "team-000", nil),
			}
		default:
			t.Errorf("unexpected skip: %d", skip)
		}

		w.Header().Set("Content-Type", "application/json")

		err = json.NewEncoder(w).Encode(map[string]any{
			"sys":   map[string]any{"type": "Array"},
			"total": 101,
			"skip":  skip,
			"limit": 100,
			"items": items,
		})
		if err != nil {
			t.Errorf("encode response: %v", err)
		}
	})

	testAccMockedResource(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id": config.StringVariable(organizationID),
				},
				// The data source sorts by team_id; positions below verify that contract.
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.contentful_teams.test", tfjsonpath.New("teams").AtSliceIndex(0).AtMapKey("description"), knownvalue.Null()),
					statecheck.ExpectKnownValue("data.contentful_teams.test", tfjsonpath.New("teams"), knownvalue.ListSizeExact(101)),
					statecheck.ExpectKnownValue("data.contentful_teams.test", tfjsonpath.New("teams").AtSliceIndex(0).AtMapKey("team_id"), knownvalue.StringExact("team-000")),
					statecheck.ExpectKnownValue("data.contentful_teams.test", tfjsonpath.New("teams").AtSliceIndex(100).AtMapKey("team_id"), knownvalue.StringExact("team-100")),
				},
				Check: func(*terraform.State) error {
					if actual := requestCount.Load(); actual != 2 {
						return fmt.Errorf("%w: expected 2, got %d", errUnexpectedTeamListRequestCount, actual)
					}

					return nil
				},
			},
		},
	})
}

func TestAccTeamsDataSourceAPIError(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		err := cmt.WriteContentfulManagementErrorResponse(
			w,
			http.StatusBadRequest,
			"BadRequest",
			new("Injected team-list failure"),
			nil,
		)
		if err != nil {
			t.Errorf("write error response: %v", err)
		}
	})

	testAccMockedResource(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id": config.StringVariable("2zuSjSO4A0e6GKBrhJRe2m"),
				},
				ExpectError: regexp.MustCompile(`(?s)Failed to read teams.*BadRequest: Injected team-list failure`),
			},
		},
	})
}

func TestAccTeamsDataSourcePaginationWithoutTotal(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int64

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)

		skip, err := strconv.Atoi(r.URL.Query().Get("skip"))
		if err != nil {
			t.Errorf("invalid skip: %v", err)
		}

		items := []map[string]any{}
		if skip == 0 {
			items = append(items, teamListItem("organization-id", "team-id", "Team", "Description"))
		} else if skip != 1 {
			t.Errorf("unexpected skip: %d", skip)
		}

		w.Header().Set("Content-Type", "application/json")

		err = json.NewEncoder(w).Encode(map[string]any{
			"sys":   map[string]any{"type": "Array"},
			"items": items,
		})
		if err != nil {
			t.Errorf("encode response: %v", err)
		}
	})

	testAccMockedResource(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id": config.StringVariable("organization-id"),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.contentful_teams.test", tfjsonpath.New("teams"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue("data.contentful_teams.test", tfjsonpath.New("teams").AtSliceIndex(0).AtMapKey("team_id"), knownvalue.StringExact("team-id")),
				},
				Check: func(*terraform.State) error {
					if actual := requestCount.Load(); actual != 2 {
						return fmt.Errorf("%w: expected 2, got %d", errUnexpectedTeamListRequestCount, actual)
					}

					return nil
				},
			},
		},
	})
}

func TestAccTeamsDataSourceAssignment(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	organizationID := "2zuSjSO4A0e6GKBrhJRe2m"

	server.RegisterSpaceEnvironment("space-id", "master")
	server.SetTeam(organizationID, "team-id", cm.TeamData{
		Name:        "SCIM Managed Team",
		Description: cm.NewNilString("Managed outside Terraform."),
	})
	server.SetTeam(organizationID, "unrelated-team-a", cm.TeamData{
		Name: "Duplicate Unrelated Team",
	})
	server.SetTeam(organizationID, "unrelated-team-b", cm.TeamData{
		Name: "Duplicate Unrelated Team",
	})

	testAccMockedResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id": config.StringVariable(organizationID),
					"space_id":        config.StringVariable("space-id"),
					"team_name":       config.StringVariable("SCIM Managed Team"),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_team_space_membership.test", tfjsonpath.New("team_id"), knownvalue.StringExact("team-id")),
					statecheck.ExpectKnownValue("contentful_team_space_membership.test", tfjsonpath.New("space_id"), knownvalue.StringExact("space-id")),
				},
			},
		},
	})
}

func teamListItem(organizationID, teamID, name string, description any) map[string]any {
	return map[string]any{
		"sys": map[string]any{
			"type": "Team",
			"id":   teamID,
			"organization": map[string]any{
				"sys": map[string]any{
					"type":     "Link",
					"linkType": "Organization",
					"id":       organizationID,
				},
			},
		},
		"name":        name,
		"description": description,
	}
}
