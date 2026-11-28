package cmtesting

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetTag(ctx context.Context, params cm.GetTagParams) (cm.GetTagRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environmentID := ts.requestEnvironmentID(ctx, params.SpaceID, params.EnvironmentID)

	tag := ts.tags.Get(params.SpaceID, environmentID, params.TagID)
	if tag == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Tag not found"), nil)), nil
	}

	response := projectTagResponse(*tag, params.EnvironmentID)

	return snapshotResponse(&response), nil
}

//nolint:ireturn
func (ts *Handler) PutTag(ctx context.Context, req *cm.TagRequest, params cm.PutTagParams) (cm.PutTagRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environmentID := ts.requestEnvironmentID(ctx, params.SpaceID, params.EnvironmentID)

	if ts.environments.Get(params.SpaceID, environmentID) == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil)), nil
	}

	tag := ts.tags.Get(params.SpaceID, environmentID, params.TagID)
	if tag == nil {
		newTag := NewTagFromRequest(params.SpaceID, environmentID, params.TagID, req)
		ts.tags.Set(params.SpaceID, environmentID, params.TagID, &newTag)

		return snapshotResponse(&cm.TagStatusCode{
			StatusCode: http.StatusCreated,
			Response:   projectTagResponse(newTag, params.EnvironmentID),
		}), nil
	}

	version, versionSet := params.XContentfulVersion.Get()
	if !versionSet || version != tag.Sys.Version {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil)), nil
	}

	UpdateTagFromRequest(tag, req)

	return snapshotResponse(&cm.TagStatusCode{
		StatusCode: http.StatusOK,
		Response:   projectTagResponse(*tag, params.EnvironmentID),
	}), nil
}

//nolint:ireturn
func (ts *Handler) DeleteTag(ctx context.Context, params cm.DeleteTagParams) (cm.DeleteTagRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environmentID := ts.requestEnvironmentID(ctx, params.SpaceID, params.EnvironmentID)

	tag := ts.tags.Get(params.SpaceID, environmentID, params.TagID)
	if tag == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Tag not found"), nil)), nil
	}

	ts.tags.Delete(params.SpaceID, environmentID, params.TagID)

	return snapshotResponse(&cm.NoContent{}), nil
}

func projectTagResponse(tag cm.Tag, environmentID string) cm.Tag {
	tag.Sys.Environment.Sys.ID = environmentID

	return tag
}
