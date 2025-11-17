package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) CreateTag(_ context.Context, req *cm.TagData, params cm.CreateTagParams) (cm.CreateTagRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.EnvironmentID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	tagID := generateResourceID()

	newTag := NewTagFromRequest(params.SpaceID, params.EnvironmentID, tagID, req, params.XContentfulTagVisibility)
	ts.tags.Set(params.SpaceID, params.EnvironmentID, tagID, &newTag)

	return &cm.TagStatusCode{
		StatusCode: http.StatusCreated,
		Response:   newTag,
	}, nil
}

//nolint:ireturn
func (ts *Handler) GetTag(_ context.Context, params cm.GetTagParams) (cm.GetTagRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.EnvironmentID == NonexistentID || params.TagID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	tag := ts.tags.Get(params.SpaceID, params.EnvironmentID, params.TagID)
	if tag == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Tag not found"), nil), nil
	}

	return tag, nil
}

//nolint:ireturn
func (ts *Handler) PutTag(_ context.Context, req *cm.TagData, params cm.PutTagParams) (cm.PutTagRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.EnvironmentID == NonexistentID || params.TagID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	existing := ts.tags.Get(params.SpaceID, params.EnvironmentID, params.TagID)
	if existing == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Tag not found"), nil), nil
	}

	updatedTag := NewTagFromRequest(params.SpaceID, params.EnvironmentID, params.TagID, req, cm.OptString{}) // Visibility not updated
	updatedTag.Sys.Version = existing.Sys.Version + 1
	ts.tags.Set(params.SpaceID, params.EnvironmentID, params.TagID, &updatedTag)

	return &cm.TagStatusCode{
		StatusCode: http.StatusOK,
		Response:   updatedTag,
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteTag(_ context.Context, params cm.DeleteTagParams) (cm.DeleteTagRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.EnvironmentID == NonexistentID || params.TagID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	tag := ts.tags.Get(params.SpaceID, params.EnvironmentID, params.TagID)
	if tag == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Tag not found"), nil), nil
	}

	ts.tags.Delete(params.SpaceID, params.EnvironmentID, params.TagID)

	return &cm.NoContent{}, nil
}
