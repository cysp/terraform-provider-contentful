package integration_tests_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTeamAcceptsJSONContentTypes(t *testing.T) {
	t.Parallel()

	for _, contentType := range []string{"application/json", "application/vnd.contentful.management.v1+json"} {
		t.Run(contentType, func(t *testing.T) {
			t.Parallel()

			team := cmt.NewTeamFromFields("organization-id", "team-id", cm.TeamData{
				Name:        "Test Team",
				Description: cm.NewNilString(""),
			})

			testserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/organizations/organization-id/teams", r.URL.Path)

				w.Header().Set("Content-Type", contentType)
				w.WriteHeader(http.StatusCreated)
				assert.NoError(t, json.NewEncoder(w).Encode(team))
			}))
			defer testserver.Close()

			client := testContentfulManagementClient(t, testserver.URL, cmt.ValidAccessToken)

			response, err := client.CreateTeam(t.Context(), &cm.TeamData{
				Name:        "Test Team",
				Description: cm.NewNilString(""),
			}, cm.CreateTeamParams{
				OrganizationID: "organization-id",
			})
			require.NoError(t, err)

			teamResponse, ok := response.(*cm.TeamStatusCode)
			require.True(t, ok)
			require.Equal(t, http.StatusCreated, teamResponse.StatusCode)
			require.Equal(t, "team-id", teamResponse.Response.Sys.ID)
		})
	}
}

func TestGetTeamAcceptsJSONContentTypes(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		contentType string
	}{
		"application/json": {
			contentType: "application/json",
		},
		"application/vnd.contentful.management.v1+json": {
			contentType: "application/vnd.contentful.management.v1+json",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			team := cmt.NewTeamFromFields("organization-id", "team-id", cm.TeamData{
				Name:        "Test Team",
				Description: cm.NewNilString(""),
			})

			testserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/organizations/organization-id/teams/team-id", r.URL.Path)

				w.Header().Set("Content-Type", test.contentType)
				assert.NoError(t, json.NewEncoder(w).Encode(team))
			}))
			defer testserver.Close()

			client := testContentfulManagementClient(t, testserver.URL, cmt.ValidAccessToken)

			response, err := client.GetTeam(t.Context(), cm.GetTeamParams{
				OrganizationID: "organization-id",
				TeamID:         "team-id",
			})
			require.NoError(t, err)

			teamResponse, ok := response.(*cm.Team)
			require.True(t, ok)
			require.Equal(t, "team-id", teamResponse.Sys.ID)
		})
	}
}

func TestGetTeamsAcceptsListItemsWithoutVersion(t *testing.T) {
	t.Parallel()

	testserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	defer testserver.Close()

	client := testContentfulManagementClient(t, testserver.URL, cmt.ValidAccessToken)

	response, err := client.GetTeams(t.Context(), cm.GetTeamsParams{
		OrganizationID: "organization-id",
		Skip:           cm.NewOptInt64(0),
		Limit:          cm.NewOptInt64(100),
	})
	require.NoError(t, err)

	teams, ok := response.(*cm.TeamCollection)
	require.True(t, ok)
	require.Len(t, teams.Items, 1)
	require.Equal(t, "team-id", teams.Items[0].Sys.ID)
	require.Equal(t, "Test Team", teams.Items[0].Name)
}

func TestGetTeamsAcceptsCollectionWithoutPaginationMetadata(t *testing.T) {
	t.Parallel()

	testserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"sys":   map[string]any{"type": "Array"},
			"items": []any{},
		}))
	}))
	defer testserver.Close()

	client := testContentfulManagementClient(t, testserver.URL, cmt.ValidAccessToken)

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

func TestPutTeamAcceptsJSONContentTypes(t *testing.T) {
	t.Parallel()

	for _, contentType := range []string{"application/json", "application/vnd.contentful.management.v1+json"} {
		t.Run(contentType, func(t *testing.T) {
			t.Parallel()

			team := cmt.NewTeamFromFields("organization-id", "team-id", cm.TeamData{
				Name:        "Test Team",
				Description: cm.NewNilString(""),
			})

			testserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Equal(t, "/organizations/organization-id/teams/team-id", r.URL.Path)
				assert.Equal(t, "7", r.Header.Get("X-Contentful-Version"))

				w.Header().Set("Content-Type", contentType)
				assert.NoError(t, json.NewEncoder(w).Encode(team))
			}))
			defer testserver.Close()

			client := testContentfulManagementClient(t, testserver.URL, cmt.ValidAccessToken)

			response, err := client.PutTeam(t.Context(), &cm.TeamData{
				Name:        "Test Team",
				Description: cm.NewNilString(""),
			}, cm.PutTeamParams{
				OrganizationID:     "organization-id",
				TeamID:             "team-id",
				XContentfulVersion: 7,
			})
			require.NoError(t, err)

			teamResponse, ok := response.(*cm.TeamStatusCode)
			require.True(t, ok)
			require.Equal(t, http.StatusOK, teamResponse.StatusCode)
			require.Equal(t, "team-id", teamResponse.Response.Sys.ID)
		})
	}
}
