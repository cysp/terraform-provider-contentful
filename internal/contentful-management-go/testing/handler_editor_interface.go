package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetEditorInterface(_ context.Context, params cm.GetEditorInterfaceParams) (cm.GetEditorInterfaceRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	editorInterface := ts.editorInterfaces.Get(params.SpaceID, params.EnvironmentID, params.ContentTypeID)
	if editorInterface == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("EditorInterface not found"), nil), nil
	}

	return editorInterface, nil
}

//nolint:ireturn
func (ts *Handler) PutEditorInterface(_ context.Context, req *cm.EditorInterfaceData, params cm.PutEditorInterfaceParams) (cm.PutEditorInterfaceRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	contentType := ts.contentTypes.Get(params.SpaceID, params.EnvironmentID, params.ContentTypeID)
	if contentType == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("ContentType not found"), nil), nil
	}

	editorInterface := ts.editorInterfaces.Get(params.SpaceID, params.EnvironmentID, params.ContentTypeID)
	if editorInterface == nil {
		newEditorInterface := NewEditorInterfaceFromFields(params.SpaceID, params.EnvironmentID, params.ContentTypeID, *req)
		ts.editorInterfaces.Set(params.SpaceID, params.EnvironmentID, params.ContentTypeID, &newEditorInterface)

		return &cm.EditorInterfaceStatusCode{
			StatusCode: http.StatusCreated,
			Response:   newEditorInterface,
		}, nil
	}

	if params.XContentfulVersion != editorInterface.Sys.Version {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	UpdateEditorInterfaceFromFields(editorInterface, *req)

	return &cm.EditorInterfaceStatusCode{
		StatusCode: http.StatusOK,
		Response:   *editorInterface,
	}, nil
}
