package testing

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
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Tag not found"), nil), nil
	}

	return tag, nil
}

//nolint:ireturn
func (ts *Handler) PutTag(_ context.Context, req *cm.TagRequest, params cm.PutTagParams) (cm.PutTagRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, params.EnvironmentID) == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Environment not found"), nil), nil
	}

	tag := ts.tags.Get(params.SpaceID, params.EnvironmentID, params.TagID)
	if tag == nil {
		newTag := NewTagFromRequest(params.SpaceID, params.EnvironmentID, params.TagID, req)
		ts.tags.Set(params.SpaceID, params.EnvironmentID, params.TagID, &newTag)

		return &cm.TagStatusCode{
			StatusCode: http.StatusCreated,
			Response:   newTag,
		}, nil
	}

	version, versionSet := params.XContentfulVersion.Get()
	if !versionSet || version != tag.Sys.Version {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	UpdateTagFromRequest(tag, req)

	return &cm.TagStatusCode{
		StatusCode: http.StatusOK,
		Response:   *tag,
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteTag(_ context.Context, params cm.DeleteTagParams) (cm.DeleteTagRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	tag := ts.tags.Get(params.SpaceID, params.EnvironmentID, params.TagID)
	if tag == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Tag not found"), nil), nil
	}

	ts.tags.Delete(params.SpaceID, params.EnvironmentID, params.TagID)

	return &cm.NoContent{}, nil
}
