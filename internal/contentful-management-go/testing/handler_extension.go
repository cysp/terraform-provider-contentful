//nolint:dupl
package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetExtension(_ context.Context, params cm.GetExtensionParams) (cm.GetExtensionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.EnvironmentID == NonexistentID || params.ExtensionID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	extension := ts.extensions.Get(params.SpaceID, params.EnvironmentID, params.ExtensionID)
	if extension == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Extension not found"), nil), nil
	}

	return extension, nil
}

//nolint:ireturn
func (ts *Handler) PutExtension(_ context.Context, req *cm.ExtensionData, params cm.PutExtensionParams) (cm.PutExtensionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.EnvironmentID == NonexistentID || params.ExtensionID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	extension := ts.extensions.Get(params.SpaceID, params.EnvironmentID, params.ExtensionID)
	if extension == nil {
		newExtension := NewExtensionFromFields(params.SpaceID, params.EnvironmentID, params.ExtensionID, *req)
		ts.extensions.Set(params.SpaceID, params.EnvironmentID, params.ExtensionID, &newExtension)

		return &cm.ExtensionStatusCode{
			StatusCode: http.StatusCreated,
			Response:   newExtension,
		}, nil
	}

	UpdateExtensionFromFields(extension, *req)

	return &cm.ExtensionStatusCode{
		StatusCode: http.StatusOK,
		Response:   *extension,
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteExtension(_ context.Context, params cm.DeleteExtensionParams) (cm.DeleteExtensionRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.EnvironmentID == NonexistentID || params.ExtensionID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	extension := ts.extensions.Get(params.SpaceID, params.EnvironmentID, params.ExtensionID)
	if extension == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Extension not found"), nil), nil
	}

	ts.extensions.Delete(params.SpaceID, params.EnvironmentID, params.ExtensionID)

	return &cm.NoContent{}, nil
}
