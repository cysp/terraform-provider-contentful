package cmtesting

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) CreatePreviewEnvironment(_ context.Context, req *cm.PreviewEnvironmentCreateData, params cm.CreatePreviewEnvironmentParams) (cm.CreatePreviewEnvironmentRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, "master") == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Space not found"), nil), nil
	}

	data := cm.NewPreviewEnvironmentDataFromCreate(*req)
	if validationError := validatePreviewEnvironmentData(data); validationError != nil {
		return validationError, nil
	}

	previewEnvironment := NewPreviewEnvironmentFromData(params.SpaceID, generateResourceID(), data)
	ts.previewEnvironments.Set(params.SpaceID, previewEnvironment.Sys.ID, &previewEnvironment)

	return &previewEnvironment, nil
}

//nolint:ireturn
func (ts *Handler) GetPreviewEnvironment(_ context.Context, params cm.GetPreviewEnvironmentParams) (cm.GetPreviewEnvironmentRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	previewEnvironment := ts.previewEnvironments.Get(params.SpaceID, params.PreviewEnvironmentID)
	if previewEnvironment == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("PreviewEnvironment not found"), nil), nil
	}

	return previewEnvironment, nil
}

//nolint:ireturn
func (ts *Handler) GetPreviewEnvironments(_ context.Context, params cm.GetPreviewEnvironmentsParams) (cm.GetPreviewEnvironmentsRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ids := ts.previewEnvironments.IDs(params.SpaceID)
	total := len(ids)

	skip := 0
	if value, ok := params.Skip.Get(); ok && value > 0 {
		skip = min(int(value), total)
	}

	limit := total
	if value, ok := params.Limit.Get(); ok && value >= 0 {
		limit = int(value)
	}

	end := min(skip+limit, total)

	items := make([]cm.PreviewEnvironment, 0, end-skip)
	for _, id := range ids[skip:end] {
		if previewEnvironment := ts.previewEnvironments.Get(params.SpaceID, id); previewEnvironment != nil {
			items = append(items, *previewEnvironment)
		}
	}

	return &cm.PreviewEnvironmentCollection{
		Sys: cm.PreviewEnvironmentCollectionSys{
			Type: cm.PreviewEnvironmentCollectionSysTypeArray,
		},
		Total: total,
		Skip:  skip,
		Limit: limit,
		Items: items,
	}, nil
}

//nolint:ireturn
func (ts *Handler) PutPreviewEnvironment(_ context.Context, req *cm.PreviewEnvironmentData, params cm.PutPreviewEnvironmentParams) (cm.PutPreviewEnvironmentRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, "master") == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Space not found"), nil), nil
	}

	if validationError := validatePreviewEnvironmentData(*req); validationError != nil {
		return validationError, nil
	}

	previewEnvironment := ts.previewEnvironments.Get(params.SpaceID, params.PreviewEnvironmentID)
	if previewEnvironment == nil {
		if params.XContentfulVersion != 0 {
			return NewContentfulManagementErrorStatusCode(http.StatusConflict, "Conflict", new("Version mismatch"), nil), nil
		}

		created := NewPreviewEnvironmentFromData(params.SpaceID, params.PreviewEnvironmentID, *req)
		ts.previewEnvironments.Set(params.SpaceID, params.PreviewEnvironmentID, &created)

		return &created, nil
	}

	if params.XContentfulVersion != previewEnvironment.Sys.Version {
		return NewContentfulManagementErrorStatusCode(http.StatusConflict, "Conflict", new("Version mismatch"), nil), nil
	}

	UpdatePreviewEnvironmentFromData(previewEnvironment, *req)

	return previewEnvironment, nil
}

//nolint:ireturn
func (ts *Handler) DeletePreviewEnvironment(_ context.Context, params cm.DeletePreviewEnvironmentParams) (cm.DeletePreviewEnvironmentRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.previewEnvironments.Get(params.SpaceID, params.PreviewEnvironmentID) == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("PreviewEnvironment not found"), nil), nil
	}

	ts.previewEnvironments.Delete(params.SpaceID, params.PreviewEnvironmentID)

	return &cm.NoContent{}, nil
}
