package cmtesting

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetTag(_ context.Context, params cm.GetTagParams) (cm.GetTagRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	tag := ts.tags.Get(params.SpaceID, params.EnvironmentID, params.TagID)
	if tag == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Tag not found"), nil)), nil
	}

	return snapshotResponse(tag), nil
}

//nolint:ireturn
func (ts *Handler) PutTag(_ context.Context, req *cm.TagRequest, params cm.PutTagParams) (cm.PutTagRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, params.EnvironmentID) == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil)), nil
	}

	tag := ts.tags.Get(params.SpaceID, params.EnvironmentID, params.TagID)
	if tag == nil {
		newTag := NewTagFromRequest(params.SpaceID, params.EnvironmentID, params.TagID, req)
		ts.tags.Set(params.SpaceID, params.EnvironmentID, params.TagID, &newTag)

		return snapshotResponse(&cm.TagStatusCode{
			StatusCode: http.StatusCreated,
			Response:   newTag,
		}), nil
	}

	version, versionSet := params.XContentfulVersion.Get()
	if !versionSet || version != tag.Sys.Version {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil)), nil
	}

	UpdateTagFromRequest(tag, req)

	return snapshotResponse(&cm.TagStatusCode{
		StatusCode: http.StatusOK,
		Response:   *tag,
	}), nil
}

//nolint:ireturn
func (ts *Handler) DeleteTag(_ context.Context, params cm.DeleteTagParams) (cm.DeleteTagRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	tag := ts.tags.Get(params.SpaceID, params.EnvironmentID, params.TagID)
	if tag == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Tag not found"), nil)), nil
	}

	ts.tags.Delete(params.SpaceID, params.EnvironmentID, params.TagID)

	return snapshotResponse(&cm.NoContent{}), nil
}
