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
		assertType  func(t *testing.T, response cm.GetTeamRes)
	}{
		"application/json": {
			contentType: "application/json",
			assertType: func(t *testing.T, response cm.GetTeamRes) {
				t.Helper()

				_, ok := response.(*cm.GetTeamApplicationJSONOK)
				require.True(t, ok)
			},
		},
		"application/vnd.contentful.management.v1+json": {
			contentType: "application/vnd.contentful.management.v1+json",
			assertType: func(t *testing.T, response cm.GetTeamRes) {
				t.Helper()

				_, ok := response.(*cm.GetTeamApplicationVndContentfulManagementV1JSONOK)
				require.True(t, ok)
			},
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

			test.assertType(t, response)

			teamResponse, ok := cm.TeamFromGetTeamResponse(response)
			require.True(t, ok)
			require.Equal(t, "team-id", teamResponse.Sys.ID)
		})
	}
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
