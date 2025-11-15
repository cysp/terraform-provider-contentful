package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) CreateTeamSpaceMembership(_ context.Context, req *cm.TeamSpaceMembershipData, params cm.CreateTeamSpaceMembershipParams) (cm.CreateTeamSpaceMembershipRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	teamSpaceMembershipID := generateResourceID()
	teamSpaceMembership := NewTeamSpaceMembershipFromFields(params.SpaceID, teamSpaceMembershipID, params.XContentfulTeam, *req)
	ts.teamSpaceMemberships.Set(params.SpaceID, teamSpaceMembership.Sys.ID, &teamSpaceMembership)

	return &cm.TeamSpaceMembershipStatusCode{
		StatusCode: http.StatusCreated,
		Response:   teamSpaceMembership,
	}, nil
}

//nolint:ireturn
func (ts *Handler) GetTeamSpaceMembership(_ context.Context, params cm.GetTeamSpaceMembershipParams) (cm.GetTeamSpaceMembershipRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.TeamSpaceMembershipID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	teamSpaceMembership := ts.teamSpaceMemberships.Get(params.SpaceID, params.TeamSpaceMembershipID)
	if teamSpaceMembership == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Team space membership not found"), nil), nil
	}

	return teamSpaceMembership, nil
}

//nolint:ireturn
func (ts *Handler) PutTeamSpaceMembership(_ context.Context, req *cm.TeamSpaceMembershipData, params cm.PutTeamSpaceMembershipParams) (cm.PutTeamSpaceMembershipRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.TeamSpaceMembershipID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	teamSpaceMembership := ts.teamSpaceMemberships.Get(params.SpaceID, params.TeamSpaceMembershipID)
	if teamSpaceMembership == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Team space membership not found"), nil), nil
	}

	UpdateTeamSpaceMembershipFromFields(teamSpaceMembership, params.SpaceID, params.TeamSpaceMembershipID, teamSpaceMembership.Sys.Team.Sys.ID, *req)

	return &cm.TeamSpaceMembershipStatusCode{
		StatusCode: http.StatusOK,
		Response:   *teamSpaceMembership,
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteTeamSpaceMembership(_ context.Context, params cm.DeleteTeamSpaceMembershipParams) (cm.DeleteTeamSpaceMembershipRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.TeamSpaceMembershipID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	teamSpaceMembership := ts.teamSpaceMemberships.Get(params.SpaceID, params.TeamSpaceMembershipID)
	if teamSpaceMembership == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Team space membership not found"), nil), nil
	}

	ts.teamSpaceMemberships.Delete(params.SpaceID, params.TeamSpaceMembershipID)

	return &cm.NoContent{}, nil
}
