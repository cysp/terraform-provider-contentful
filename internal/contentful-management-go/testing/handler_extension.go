package cmtesting

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetExtension(ctx context.Context, params cm.GetExtensionParams) (cm.GetExtensionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environmentID := ts.requestEnvironmentID(ctx, params.SpaceID, params.EnvironmentID)

	extension := ts.extensions.Get(params.SpaceID, environmentID, params.ExtensionID)
	if extension == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Extension not found"), nil)), nil
	}

	response := projectExtensionResponse(*extension, params.EnvironmentID)

	return snapshotResponse(&response), nil
}

//nolint:ireturn
func (ts *Handler) PutExtension(ctx context.Context, req *cm.ExtensionData, params cm.PutExtensionParams) (cm.PutExtensionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environmentID := ts.requestEnvironmentID(ctx, params.SpaceID, params.EnvironmentID)

	if ts.environments.Get(params.SpaceID, environmentID) == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil)), nil
	}

	extension := ts.extensions.Get(params.SpaceID, environmentID, params.ExtensionID)
	if extension == nil {
		newExtension := NewExtensionFromFields(params.SpaceID, environmentID, params.ExtensionID, *req)
		ts.extensions.Set(params.SpaceID, environmentID, params.ExtensionID, &newExtension)

		return snapshotResponse(&cm.ExtensionStatusCode{
			StatusCode: http.StatusCreated,
			Response:   projectExtensionResponse(newExtension, params.EnvironmentID),
		}), nil
	}

	if params.XContentfulVersion != extension.Sys.Version {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	UpdateExtensionFromFields(extension, *req)

	return snapshotResponse(&cm.ExtensionStatusCode{
		StatusCode: http.StatusOK,
		Response:   projectExtensionResponse(*extension, params.EnvironmentID),
	}), nil
}

//nolint:ireturn
func (ts *Handler) DeleteExtension(ctx context.Context, params cm.DeleteExtensionParams) (cm.DeleteExtensionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environmentID := ts.requestEnvironmentID(ctx, params.SpaceID, params.EnvironmentID)

	extension := ts.extensions.Get(params.SpaceID, environmentID, params.ExtensionID)
	if extension == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Extension not found"), nil)), nil
	}

	ts.extensions.Delete(params.SpaceID, environmentID, params.ExtensionID)

	return snapshotResponse(&cm.NoContent{}), nil
}

func projectExtensionResponse(extension cm.Extension, environmentID string) cm.Extension {
	extension.Sys.Environment.Sys.ID = environmentID

	return extension
}
