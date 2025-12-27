package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetTags(_ context.Context, params cm.GetTagsParams) (cm.GetTagsRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.EnvironmentID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	tags := ts.tags.List(params.SpaceID, params.EnvironmentID)

	items := make([]cm.Tag, 0, len(tags))
	for _, tag := range tags {
		items = append(items, *tag)
	}

	return &cm.TagCollection{
		Sys: cm.TagCollectionSys{
			Type: cm.TagCollectionSysTypeArray,
		},
		Total: cm.NewOptInt(len(tags)),
		Items: items,
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
func (ts *Handler) PutTag(_ context.Context, req *cm.TagRequest, params cm.PutTagParams) (cm.PutTagRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.EnvironmentID == NonexistentID || params.TagID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
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
