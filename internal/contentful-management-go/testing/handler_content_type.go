package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetContentTypes(_ context.Context, params cm.GetContentTypesParams) (cm.GetContentTypesRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, params.EnvironmentID) == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Environment not found"), nil), nil
	}

	contentTypes := ts.contentTypes.List(params.SpaceID, params.EnvironmentID)

	items := make([]cm.ContentType, 0, len(contentTypes))
	for _, ct := range contentTypes {
		items = append(items, *ct)
	}

	return &cm.ContentTypeCollection{
		Sys: cm.ContentTypeCollectionSys{
			Type: cm.ContentTypeCollectionSysTypeArray,
		},
		Total: cm.NewOptInt(len(contentTypes)),
		Items: items,
	}, nil
}

//nolint:ireturn
func (ts *Handler) GetContentType(_ context.Context, params cm.GetContentTypeParams) (cm.GetContentTypeRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	contentType := ts.contentTypes.Get(params.SpaceID, params.EnvironmentID, params.ContentTypeID)
	if contentType == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("ContentType not found"), nil), nil
	}

	return contentType, nil
}

//nolint:ireturn
func (ts *Handler) PutContentType(_ context.Context, req *cm.ContentTypeRequestData, params cm.PutContentTypeParams) (cm.PutContentTypeRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environment := ts.environments.Get(params.SpaceID, params.EnvironmentID)
	if environment == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Environment not found"), nil), nil
	}

	contentType := ts.contentTypes.Get(params.SpaceID, params.EnvironmentID, params.ContentTypeID)
	if contentType == nil {
		newContentType := NewContentTypeFromRequestFields(params.SpaceID, params.EnvironmentID, params.ContentTypeID, *req)
		ts.contentTypes.Set(params.SpaceID, params.EnvironmentID, params.ContentTypeID, &newContentType)

		return &cm.ContentTypeStatusCode{
			StatusCode: http.StatusCreated,
			Response:   newContentType,
		}, nil
	}

	if params.XContentfulVersion != contentType.Sys.Version {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	UpdateContentTypeFromRequestFields(contentType, *req)

	editorInterface := ts.editorInterfaces.Get(params.SpaceID, params.EnvironmentID, params.ContentTypeID)
	if editorInterface != nil {
		editorInterface.Sys.Version++
	}

	return &cm.ContentTypeStatusCode{
		StatusCode: http.StatusOK,
		Response:   *contentType,
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteContentType(_ context.Context, params cm.DeleteContentTypeParams) (cm.DeleteContentTypeRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	contentType := ts.contentTypes.Get(params.SpaceID, params.EnvironmentID, params.ContentTypeID)
	if contentType == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("ContentType not found"), nil), nil
	}

	ts.contentTypes.Delete(params.SpaceID, params.EnvironmentID, params.ContentTypeID)

	return &cm.NoContent{}, nil
}

//nolint:ireturn
func (ts *Handler) ActivateContentType(_ context.Context, params cm.ActivateContentTypeParams) (cm.ActivateContentTypeRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	contentType := ts.contentTypes.Get(params.SpaceID, params.EnvironmentID, params.ContentTypeID)
	if contentType == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("ContentType not found"), nil), nil
	}

	publishContentType(contentType)

	return &cm.ContentTypeStatusCode{
		StatusCode: http.StatusOK,
		Response:   *contentType,
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeactivateContentType(_ context.Context, params cm.DeactivateContentTypeParams) (cm.DeactivateContentTypeRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	contentType := ts.contentTypes.Get(params.SpaceID, params.EnvironmentID, params.ContentTypeID)
	if contentType == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("ContentType not found"), nil), nil
	}

	contentType.Sys.PublishedVersion.Reset()

	return &cm.NoContent{}, nil
}
