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
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var errUnexpectedTeamListRequestCount = errors.New("unexpected team-list request count")

// These acceptance tests deliberately use the mocked Contentful service. They
// validate Terraform behavior and the subset of the published HTTP contract
// consumed by the provider, not live authorization or visibility of
// SCIM-managed teams.

func TestAccTeamsDataSource(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()
	organizationID := "2zuSjSO4A0e6GKBrhJRe2m"

	server.SetTeam(organizationID, "team-b", cm.TeamData{
		Name:        "Second Team",
		Description: cm.NewNilString("The second team."),
	})
	server.SetTeam(organizationID, "team-a", cm.TeamData{
		Name:        "First Team",
		Description: cm.NewNilString("The first team."),
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id": config.StringVariable(organizationID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.contentful_teams.test", "id", organizationID),
					resource.TestCheckResourceAttr("data.contentful_teams.test", "organization_id", organizationID),
					resource.TestCheckResourceAttr("data.contentful_teams.test", "teams.#", "2"),
					resource.TestCheckResourceAttr("data.contentful_teams.test", "teams.0.team_id", "team-a"),
					resource.TestCheckResourceAttr("data.contentful_teams.test", "teams.0.name", "First Team"),
					resource.TestCheckResourceAttr("data.contentful_teams.test", "teams.0.description", "The first team."),
					resource.TestCheckResourceAttr("data.contentful_teams.test", "teams.1.team_id", "team-b"),
					resource.TestCheckResourceAttr("data.contentful_teams.test", "teams.1.name", "Second Team"),
					resource.TestCheckResourceAttr("data.contentful_teams.test", "teams.1.description", "The second team."),
				),
			},
		},
	})
}

func TestAccTeamsDataSourceEmpty(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()
	organizationID := "2zuSjSO4A0e6GKBrhJRe2m"

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id": config.StringVariable(organizationID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.contentful_teams.test", "id", organizationID),
					resource.TestCheckResourceAttr("data.contentful_teams.test", "teams.#", "0"),
				),
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

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id": config.StringVariable(organizationID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.contentful_teams.test", "teams.#", "101"),
					resource.TestCheckResourceAttr("data.contentful_teams.test", "teams.0.team_id", "team-000"),
					resource.TestCheckNoResourceAttr("data.contentful_teams.test", "teams.0.description"),
					resource.TestCheckResourceAttr("data.contentful_teams.test", "teams.100.team_id", "team-100"),
					func(*terraform.State) error {
						if actual := requestCount.Load(); actual != 2 {
							return fmt.Errorf("%w: expected 2, got %d", errUnexpectedTeamListRequestCount, actual)
						}

						return nil
					},
				),
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

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
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

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id": config.StringVariable("organization-id"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.contentful_teams.test", "teams.#", "1"),
					resource.TestCheckResourceAttr("data.contentful_teams.test", "teams.0.team_id", "team-id"),
					func(*terraform.State) error {
						if actual := requestCount.Load(); actual != 2 {
							return fmt.Errorf("%w: expected 2, got %d", errUnexpectedTeamListRequestCount, actual)
						}

						return nil
					},
				),
			},
		},
	})
}

func TestAccTeamsDataSourceAssignment(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()
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

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id": config.StringVariable(organizationID),
					"space_id":        config.StringVariable("space-id"),
					"team_name":       config.StringVariable("SCIM Managed Team"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_team_space_membership.test", "team_id", "team-id"),
					resource.TestCheckResourceAttr("contentful_team_space_membership.test", "space_id", "space-id"),
				),
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
