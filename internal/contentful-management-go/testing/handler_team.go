package cmtesting

import (
	"cmp"
	"context"
	"net/http"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetTeams(_ context.Context, params cm.GetTeamsParams) (cm.GetTeamsRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	teams := ts.teams.Values(params.OrganizationID)
	slices.SortFunc(teams, func(a, b *cm.Team) int {
		return cmp.Compare(a.Sys.ID, b.Sys.ID)
	})

	skip := params.Skip.Or(0)
	limit := params.Limit.Or(100) //nolint:mnd

	if skip < 0 || limit < 1 || limit > 1000 {
		return NewContentfulManagementErrorStatusCodeBadRequest(new("Invalid pagination parameters"), nil), nil
	}

	start := min(skip, int64(len(teams)))
	end := min(start+limit, int64(len(teams)))

	items := make([]cm.TeamListItem, 0, end-start)
	for _, team := range teams[start:end] {
		items = append(items, cm.TeamListItem{
			Sys: cm.TeamListItemSys{
				Organization: team.Sys.Organization,
				Type:         cm.TeamListItemSysTypeTeam,
				ID:           team.Sys.ID,
			},
			Name:        team.Name,
			Description: team.Description,
		})
	}

	return &cm.TeamCollection{
		Sys: cm.TeamCollectionSys{
			Type: cm.TeamCollectionSysTypeArray,
		},
		Total: cm.NewOptInt(len(teams)),
		Skip:  cm.NewOptInt(int(start)),
		Limit: cm.NewOptInt(int(limit)),
		Items: items,
	}, nil
}

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
		return NewContentfulManagementErrorStatusCodeNotFound(new("Team not found"), nil), nil
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
		return NewContentfulManagementErrorStatusCodeNotFound(new("Team not found"), nil), nil
	}

	ts.teams.Delete(params.OrganizationID, params.TeamID)

	return &cm.NoContent{}, nil
}
