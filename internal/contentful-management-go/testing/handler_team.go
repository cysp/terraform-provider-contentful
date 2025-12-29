package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) CreateTeam(_ context.Context, req *cm.TeamData, params cm.CreateTeamParams) (cm.CreateTeamRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	teamID := generateResourceID()
	team := NewTeamFromFields(params.OrganizationID, teamID, *req)
	ts.teams.Set(params.OrganizationID, team.Sys.ID, &team)

	return &cm.TeamStatusCode{
		StatusCode: http.StatusCreated,
		Response:   team,
	}, nil
}

//nolint:ireturn
func (ts *Handler) GetTeam(_ context.Context, params cm.GetTeamParams) (cm.GetTeamRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	team := ts.teams.Get(params.OrganizationID, params.TeamID)
	if team == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Team not found"), nil), nil
	}

	return team, nil
}

//nolint:ireturn
func (ts *Handler) PutTeam(_ context.Context, req *cm.TeamData, params cm.PutTeamParams) (cm.PutTeamRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	team := ts.teams.Get(params.OrganizationID, params.TeamID)
	if team == nil {
		newTeam := NewTeamFromFields(params.OrganizationID, params.TeamID, *req)
		ts.teams.Set(params.OrganizationID, newTeam.Sys.ID, &newTeam)

		return &cm.TeamStatusCode{
			StatusCode: http.StatusCreated,
			Response:   newTeam,
		}, nil
	}

	if params.XContentfulVersion != team.Sys.Version {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	UpdateTeamFromFields(team, params.OrganizationID, params.TeamID, *req)

	return &cm.TeamStatusCode{
		StatusCode: http.StatusOK,
		Response:   *team,
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteTeam(_ context.Context, params cm.DeleteTeamParams) (cm.DeleteTeamRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	team := ts.teams.Get(params.OrganizationID, params.TeamID)
	if team == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Team not found"), nil), nil
	}

	ts.teams.Delete(params.OrganizationID, params.TeamID)

	return &cm.NoContent{}, nil
}
