package cmtesting

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) CreatePreviewEnvironment(_ context.Context, req *cm.PreviewEnvironmentData, params cm.CreatePreviewEnvironmentParams) (cm.CreatePreviewEnvironmentRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, "master") == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Space not found"), nil)), nil
	}

	if validationError := validatePreviewEnvironmentData(*req); validationError != nil {
		return snapshotResponse(validationError), nil
	}

	previewEnvironment := NewPreviewEnvironmentFromData(params.SpaceID, generateResourceID(), *req)
	ts.previewEnvironments.Set(params.SpaceID, previewEnvironment.Sys.ID, &previewEnvironment)

	return snapshotResponse(&previewEnvironment), nil
}

//nolint:ireturn
func (ts *Handler) GetPreviewEnvironment(_ context.Context, params cm.GetPreviewEnvironmentParams) (cm.GetPreviewEnvironmentRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	previewEnvironment := ts.previewEnvironments.Get(params.SpaceID, params.PreviewEnvironmentID)
	if previewEnvironment == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("PreviewEnvironment not found"), nil)), nil
	}

	return snapshotResponse(previewEnvironment), nil
}

//nolint:ireturn
func (ts *Handler) PutPreviewEnvironment(_ context.Context, req *cm.PreviewEnvironmentData, params cm.PutPreviewEnvironmentParams) (cm.PutPreviewEnvironmentRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, "master") == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Space not found"), nil)), nil
	}

	if validationError := validatePreviewEnvironmentData(*req); validationError != nil {
		return snapshotResponse(validationError), nil
	}

	previewEnvironment := ts.previewEnvironments.Get(params.SpaceID, params.PreviewEnvironmentID)
	if previewEnvironment == nil {
		if params.XContentfulVersion != 0 {
			return snapshotResponse(NewContentfulManagementErrorStatusCode(http.StatusConflict, "Conflict", new("Version mismatch"), nil)), nil
		}

		created := NewPreviewEnvironmentFromData(params.SpaceID, params.PreviewEnvironmentID, *req)
		ts.previewEnvironments.Set(params.SpaceID, params.PreviewEnvironmentID, &created)

		return snapshotResponse(&created), nil
	}

	if params.XContentfulVersion != previewEnvironment.Sys.Version {
		return snapshotResponse(NewContentfulManagementErrorStatusCode(http.StatusConflict, "Conflict", new("Version mismatch"), nil)), nil
	}

	UpdatePreviewEnvironmentFromData(previewEnvironment, *req)

	return snapshotResponse(previewEnvironment), nil
}

//nolint:ireturn
func (ts *Handler) DeletePreviewEnvironment(_ context.Context, params cm.DeletePreviewEnvironmentParams) (cm.DeletePreviewEnvironmentRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.previewEnvironments.Get(params.SpaceID, params.PreviewEnvironmentID) == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("PreviewEnvironment not found"), nil)), nil
	}

	ts.previewEnvironments.Delete(params.SpaceID, params.PreviewEnvironmentID)

	return snapshotResponse(&cm.NoContent{}), nil
}
