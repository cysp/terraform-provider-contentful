package cmtesting

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

// App requests use the target selected during authorization, even if the alias
// moves before dispatch. Other callers resolve the address under the handler lock.
func (ts *Handler) requestEnvironmentID(ctx context.Context, spaceID, environmentID string) string {
	if target, ok := ctx.Value(authorizedEnvironmentKey{}).(string); ok {
		return target
	}

	return ts.resolveEnvironmentAlias(spaceID, environmentID)
}

func (ts *Handler) resolveEnvironmentAlias(spaceID, environmentID string) string {
	if alias := ts.environmentAliases.Get(spaceID, environmentID); alias != nil {
		return alias.Environment.Sys.ID
	}

	return environmentID
}

//nolint:ireturn
func (ts *Handler) CreateOrUpdateEnvironmentAlias(_ context.Context, data *cm.EnvironmentAliasData, params cm.CreateOrUpdateEnvironmentAliasParams) (cm.CreateOrUpdateEnvironmentAliasRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environment := ts.environments.Get(params.SpaceID, data.Environment.Sys.ID)
	if environment == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil)), nil
	}

	environmentAlias := ts.environmentAliases.Get(params.SpaceID, params.EnvironmentAliasID)
	if environmentAlias == nil {
		environmentAlias := NewEnvironmentAliasFromEnvironmentAliasData(params.SpaceID, params.EnvironmentAliasID, *data)
		ts.environmentAliases.Set(params.SpaceID, params.EnvironmentAliasID, &environmentAlias)

		return snapshotResponse(&cm.EnvironmentAliasStatusCode{
			StatusCode: http.StatusCreated,
			Response:   environmentAlias,
		}), nil
	}

	if params.XContentfulVersion != environmentAlias.Sys.Version {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil)), nil
	}

	UpdateEnvironmentAliasFromEnvironmentAliasData(environmentAlias, *data)

	return snapshotResponse(&cm.EnvironmentAliasStatusCode{
		StatusCode: http.StatusOK,
		Response:   *environmentAlias,
	}), nil
}

//nolint:ireturn
func (ts *Handler) GetEnvironmentAlias(_ context.Context, params cm.GetEnvironmentAliasParams) (cm.GetEnvironmentAliasRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environmentAlias := ts.environmentAliases.Get(params.SpaceID, params.EnvironmentAliasID)
	if environmentAlias == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(nil, nil)), nil
	}

	return snapshotResponse(environmentAlias), nil
}

//nolint:ireturn
func (ts *Handler) DeleteEnvironmentAlias(_ context.Context, params cm.DeleteEnvironmentAliasParams) (cm.DeleteEnvironmentAliasRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environmentAlias := ts.environmentAliases.Get(params.SpaceID, params.EnvironmentAliasID)
	if environmentAlias == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(nil, nil)), nil
	}

	ts.environmentAliases.Delete(params.SpaceID, params.EnvironmentAliasID)

	return snapshotResponse(&cm.NoContent{}), nil
}
