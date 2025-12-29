package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) CreateOrUpdateEnvironmentAlias(_ context.Context, data *cm.EnvironmentAliasData, params cm.CreateOrUpdateEnvironmentAliasParams) (cm.CreateOrUpdateEnvironmentAliasRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environment := ts.environments.Get(params.SpaceID, data.Environment.Sys.ID)
	if environment == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Environment not found"), nil), nil
	}

	environmentAlias := ts.environmentAliases.Get(params.SpaceID, params.EnvironmentAliasID)
	if environmentAlias == nil {
		environmentAlias := NewEnvironmentAliasFromEnvironmentAliasData(params.SpaceID, params.EnvironmentAliasID, *data)
		ts.environmentAliases.Set(params.SpaceID, params.EnvironmentAliasID, &environmentAlias)

		return &cm.EnvironmentAliasStatusCode{
			StatusCode: http.StatusCreated,
			Response:   environmentAlias,
		}, nil
	}

	if params.XContentfulVersion != environmentAlias.Sys.Version {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	UpdateEnvironmentAliasFromEnvironmentAliasData(environmentAlias, *data)

	return &cm.EnvironmentAliasStatusCode{
		StatusCode: http.StatusOK,
		Response:   *environmentAlias,
	}, nil
}

//nolint:ireturn
func (ts *Handler) GetEnvironmentAlias(_ context.Context, params cm.GetEnvironmentAliasParams) (cm.GetEnvironmentAliasRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environmentAlias := ts.environmentAliases.Get(params.SpaceID, params.EnvironmentAliasID)
	if environmentAlias == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	return environmentAlias, nil
}

//nolint:ireturn
func (ts *Handler) DeleteEnvironmentAlias(_ context.Context, params cm.DeleteEnvironmentAliasParams) (cm.DeleteEnvironmentAliasRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environmentAlias := ts.environmentAliases.Get(params.SpaceID, params.EnvironmentAliasID)
	if environmentAlias == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	ts.environmentAliases.Delete(params.SpaceID, params.EnvironmentAliasID)

	return &cm.NoContent{}, nil
}
