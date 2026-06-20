package contentfulmanagement_test

import (
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/require"
)

func TestTeamFromGetTeamResponse(t *testing.T) {
	t.Parallel()

	team := cm.Team{
		Sys: cm.TeamSys{
			ID:      "team-id",
			Version: 7,
		},
	}

	tests := map[string]struct {
		response cm.GetTeamRes
		wantOK   bool
	}{
		"application json": {
			response: func() cm.GetTeamRes {
				response := cm.GetTeamApplicationJSONOK(team)

				return &response
			}(),
			wantOK: true,
		},
		"vendor json": {
			response: func() cm.GetTeamRes {
				response := cm.GetTeamApplicationVndContentfulManagementV1JSONOK(team)

				return &response
			}(),
			wantOK: true,
		},
		"non team response": {
			response: &cm.ApplicationJSONErrorStatusCode{
				StatusCode: http.StatusUnauthorized,
			},
			wantOK: false,
		},
		"nil": {
			response: nil,
			wantOK:   false,
		},
		"typed nil application json": {
			response: (*cm.GetTeamApplicationJSONOK)(nil),
			wantOK:   false,
		},
		"typed nil vendor json": {
			response: (*cm.GetTeamApplicationVndContentfulManagementV1JSONOK)(nil),
			wantOK:   false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, ok := cm.TeamFromGetTeamResponse(test.response)

			require.Equal(t, test.wantOK, ok)

			if test.wantOK {
				require.Equal(t, team, got)
			}
		})
	}
}
