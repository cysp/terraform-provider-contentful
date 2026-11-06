package integration_tests_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamClientWireContract(t *testing.T) {
	t.Parallel()

	responseBody := testJSON(map[string]any{
		"sys": map[string]any{
			"type": "Team", "id": "team-id", "version": 8,
			"organization": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Organization", "id": "organization-id"}},
		},
		"name": "Test Team", "description": "",
	})

	for name, test := range map[string]struct {
		method  string
		path    string
		status  int
		version string
	}{
		"create": {method: http.MethodPost, path: "/organizations/organization-id/teams", status: http.StatusCreated},
		"read":   {method: http.MethodGet, path: "/organizations/organization-id/teams/team-id", status: http.StatusOK},
		"update": {method: http.MethodPut, path: "/organizations/organization-id/teams/team-id", status: http.StatusOK, version: "7"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for _, contentType := range []string{"application/json", "application/vnd.contentful.management.v1+json"} {
				t.Run(contentType, func(t *testing.T) {
					t.Parallel()

					testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						assert.Equal(t, test.method, r.Method)
						assert.Equal(t, test.path, r.URL.Path)
						assert.Equal(t, test.version, r.Header.Get("X-Contentful-Version"))

						body, err := io.ReadAll(r.Body)
						assert.NoError(t, err)

						if test.method == http.MethodGet {
							assert.Empty(t, body)
						} else {
							assert.JSONEq(t, testJSON(map[string]any{"name": "Test Team", "description": ""}), string(body))
						}

						w.Header().Set("Content-Type", contentType)
						w.WriteHeader(test.status)
						_, err = io.WriteString(w, responseBody)
						assert.NoError(t, err)
					}))
					t.Cleanup(testServer.Close)

					client := testContentfulManagementClient(t, testServer.URL, cmt.ValidAccessToken)

					var (
						response any
						err      error
					)

					switch test.method {
					case http.MethodPost:
						response, err = client.CreateTeam(t.Context(), &cm.TeamData{Name: "Test Team", Description: cm.NewNilString("")}, cm.CreateTeamParams{OrganizationID: "organization-id"})
					case http.MethodGet:
						response, err = client.GetTeam(t.Context(), cm.GetTeamParams{OrganizationID: "organization-id", TeamID: "team-id"})
					case http.MethodPut:
						response, err = client.PutTeam(t.Context(), &cm.TeamData{Name: "Test Team", Description: cm.NewNilString("")}, cm.PutTeamParams{OrganizationID: "organization-id", TeamID: "team-id", XContentfulVersion: 7})
					}

					require.NoError(t, err)

					var team cm.Team

					if test.method == http.MethodGet {
						got, ok := response.(*cm.Team)
						require.True(t, ok, "expected a Team, got %T", response)
						require.NotNil(t, got)
						team = *got
					} else {
						got, ok := response.(*cm.TeamStatusCode)
						require.True(t, ok, "expected a TeamStatusCode, got %T", response)
						require.NotNil(t, got)
						assert.Equal(t, test.status, got.StatusCode)
						team = got.Response
					}

					assert.Equal(t, "team-id", team.Sys.ID)
					assert.Equal(t, 8, team.Sys.Version)
					assert.Equal(t, "Test Team", team.Name)
					assert.Equal(t, cm.NilString{Value: ""}, team.Description)
				})
			}
		})
	}
}

func TestGetTeamsAcceptsListItemsWithoutVersion(t *testing.T) {
	t.Parallel()

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/organizations/organization-id/teams", r.URL.Path)
		assert.Equal(t, "Bearer "+cmt.ValidAccessToken, r.Header.Get("Authorization"))
		assert.Equal(t, "0", r.URL.Query().Get("skip"))
		assert.Equal(t, "100", r.URL.Query().Get("limit"))

		w.Header().Set("Content-Type", "application/json")
		assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"sys": map[string]any{
				"type": "Array",
			},
			"total": 1,
			"skip":  0,
			"limit": 100,
			"items": []any{
				map[string]any{
					"sys": map[string]any{
						"type": "Team",
						"id":   "team-id",
						"organization": map[string]any{
							"sys": map[string]any{
								"type":     "Link",
								"linkType": "Organization",
								"id":       "organization-id",
							},
						},
					},
					"name":        "Test Team",
					"description": nil,
				},
			},
		}))
	}))
	t.Cleanup(testServer.Close)

	client := testContentfulManagementClient(t, testServer.URL, cmt.ValidAccessToken)

	response, err := client.GetTeams(t.Context(), cm.GetTeamsParams{
		OrganizationID: "organization-id",
		Skip:           cm.NewOptInt64(0),
		Limit:          cm.NewOptInt64(100),
	})
	require.NoError(t, err)

	teams, ok := response.(*cm.TeamCollection)
	require.True(t, ok)
	require.Len(t, teams.Items, 1)
	assert.Equal(t, "team-id", teams.Items[0].Sys.ID)
	assert.Equal(t, "Test Team", teams.Items[0].Name)
}

func TestGetTeamsAcceptsCollectionWithoutPaginationMetadata(t *testing.T) {
	t.Parallel()

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"sys":   map[string]any{"type": "Array"},
			"items": []any{},
		}))
	}))
	t.Cleanup(testServer.Close)

	client := testContentfulManagementClient(t, testServer.URL, cmt.ValidAccessToken)

	response, err := client.GetTeams(t.Context(), cm.GetTeamsParams{
		OrganizationID: "organization-id",
	})
	require.NoError(t, err)

	teams, ok := response.(*cm.TeamCollection)
	require.True(t, ok)
	assert.Empty(t, teams.Items)
	assert.False(t, teams.Total.IsSet())
	assert.False(t, teams.Skip.IsSet())
	assert.False(t, teams.Limit.IsSet())
}
