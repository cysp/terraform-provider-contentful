package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetEnvironmentAlias(_ context.Context, params cm.GetEnvironmentAliasParams) (cm.GetEnvironmentAliasRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	environmentAlias := ts.environmentAliases.Get(params.SpaceID, params.EnvironmentAliasID)
	if environmentAlias == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	return environmentAlias, nil
}

//nolint:ireturn
func (ts *Handler) CreateOrUpdateEnvironmentAlias(_ context.Context, req *cm.EnvironmentAliasRequest, params cm.CreateOrUpdateEnvironmentAliasParams) (cm.CreateOrUpdateEnvironmentAliasRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	existingEnvironmentAlias := ts.environmentAliases.Get(params.SpaceID, params.EnvironmentAliasID)

	var environmentAlias *cm.EnvironmentAlias
	var statusCode int

	if existingEnvironmentAlias == nil {
		// Create new environment alias
		newEnvironmentAlias := NewEnvironmentAliasFromRequest(params.SpaceID, params.EnvironmentAliasID, *req)
		environmentAlias = &newEnvironmentAlias
		statusCode = http.StatusCreated
	} else {
		// Update existing environment alias
		environmentAlias = existingEnvironmentAlias
		UpdateEnvironmentAliasFromRequest(environmentAlias, *req)
		statusCode = http.StatusOK
	}

	ts.environmentAliases.Set(params.SpaceID, params.EnvironmentAliasID, environmentAlias)

	return &cm.EnvironmentAliasStatusCode{
		StatusCode: statusCode,
		Response:   *environmentAlias,
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteEnvironmentAlias(_ context.Context, params cm.DeleteEnvironmentAliasParams) (cm.DeleteEnvironmentAliasRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	environmentAlias := ts.environmentAliases.Get(params.SpaceID, params.EnvironmentAliasID)
	if environmentAlias == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	ts.environmentAliases.Delete(params.SpaceID, params.EnvironmentAliasID)

	return &cm.NoContent{}, nil
}
