package contentfulmanagementtestserver

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

type ConstraintOptNil[T any] interface {
	Get() (T, bool)
	SetTo(value T)
}

func convertOptNil[I any, O any](o ConstraintOptNil[O], i ConstraintOptNil[I], f func(I) O) {
	if value, ok := i.Get(); ok {
		o.SetTo(f(value))
	}
}

func convertSlice[I any, O any](i []I, f func(I) O) []O {
	o := make([]O, len(i))
	for index, item := range i {
		o[index] = f(item)
	}

	return o
}
func convertMap[I any, O any](i map[string]I, f func(I) O) map[string]O {
	o := make(map[string]O, len(i))

	for key, item := range i {
		o[key] = f(item)
	}

	return o
}

//nolint:cyclop,gocognit
func (ts *ContentfulManagementTestServer) setupSpaceEnvironmentContentTypeHandlers() {
	ts.serveMux.Handle("/spaces/{spaceID}/environments/{environmentID}/content_types/{contentTypeID}", http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")
		environmentID := r.PathValue("environmentID")
		contentTypeID := r.PathValue("contentTypeID")

		if spaceID == NonexistentID || environmentID == NonexistentID || contentTypeID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

			return
		}

		ts.mu.Lock()
		defer ts.mu.Unlock()

		switch r.Method {
		case http.MethodGet:
			contentType, exists := ts.contentTypes.Get(spaceID, environmentID, contentTypeID)
			if !exists {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, contentType)
		case http.MethodPut:
			_, exists := ts.contentTypes.Get(spaceID, environmentID, contentTypeID)

			var contentTypeRequestFields cm.ContentTypeRequestFields
			if err := ReadContentfulManagementRequest(r, &contentTypeRequestFields); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(responseWriter, err)

				return
			}

			contentType := NewContentTypeFromRequestFields(spaceID, environmentID, contentTypeID, contentTypeRequestFields)

			ts.contentTypes.Set(spaceID, environmentID, contentTypeID, &contentType)

			statusCode := http.StatusCreated
			if exists {
				statusCode = http.StatusOK
			}

			_ = WriteContentfulManagementResponse(responseWriter, statusCode, &contentType)

		case http.MethodDelete:
			ts.contentTypes.Delete(spaceID, environmentID, contentTypeID)

			responseWriter.WriteHeader(http.StatusNoContent)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)
		}
	}))

	ts.serveMux.Handle("/spaces/{spaceID}/environments/{environmentID}/content_types/{contentTypeID}/published", http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")
		environmentID := r.PathValue("environmentID")
		contentTypeID := r.PathValue("contentTypeID")

		if spaceID == NonexistentID || environmentID == NonexistentID || contentTypeID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

			return
		}

		ts.mu.Lock()
		defer ts.mu.Unlock()

		switch r.Method {
		case http.MethodPut:
			contentType, exists := ts.contentTypes.Get(spaceID, environmentID, contentTypeID)
			if !exists {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			contentType.Sys.PublishedVersion.SetTo(contentType.Sys.Version)

			contentType.Sys.Version++

			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, contentType)

		case http.MethodDelete:
			contentType, exists := ts.contentTypes.Get(spaceID, environmentID, contentTypeID)
			if !exists {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			contentType.Sys.PublishedVersion.Reset()

			responseWriter.WriteHeader(http.StatusNoContent)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)
		}
	}))

	ts.serveMux.Handle("/spaces/{spaceID}/environments/{environmentID}/content_types/{contentTypeID}/editor_interface", http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")
		environmentID := r.PathValue("environmentID")
		contentTypeID := r.PathValue("contentTypeID")

		if spaceID == NonexistentID || environmentID == NonexistentID || contentTypeID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

			return
		}

		ts.mu.Lock()
		defer ts.mu.Unlock()

		switch r.Method {
		case http.MethodGet:
			editorInterface, exists := ts.editorInterfaces.Get(spaceID, environmentID, contentTypeID)
			if !exists {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, editorInterface)

		case http.MethodPut:
			_, exists := ts.contentTypes.Get(spaceID, environmentID, contentTypeID)
			if !exists {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			editorInterfaceFields := cm.EditorInterfaceFields{}
			if err := ReadContentfulManagementRequest(r, &editorInterfaceFields); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(responseWriter, err)

				return
			}

			editorInterface := NewEditorInterfaceFromFields(spaceID, environmentID, contentTypeID, editorInterfaceFields)

			ts.editorInterfaces.Set(spaceID, environmentID, contentTypeID, &editorInterface)

			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, &editorInterface)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)
		}
	}))
}

func (ts *ContentfulManagementTestServer) GetContentType(spaceID, environmentID, contentTypeID string) (*cm.ContentType, bool) {
	return ts.contentTypes.Get(spaceID, environmentID, contentTypeID)
}

func (ts *ContentfulManagementTestServer) SetContentType(spaceID, environmentID, contentTypeID string, contentTypeFields cm.ContentTypeRequestFields) {
	contentType := NewContentTypeFromRequestFields(spaceID, environmentID, contentTypeID, contentTypeFields)
	ts.contentTypes.Set(spaceID, environmentID, contentType.Sys.ID, &contentType)
}

func (ts *ContentfulManagementTestServer) DeleteContentType(spaceID, environmentID, contentTypeID string) {
	ts.contentTypes.Delete(spaceID, environmentID, contentTypeID)
	ts.DeleteEditorInterface(spaceID, environmentID, contentTypeID)
}

func (ts *ContentfulManagementTestServer) GetEditorInterface(spaceID, environmentID, contentTypeID string) (*cm.EditorInterface, bool) {
	return ts.editorInterfaces.Get(spaceID, environmentID, contentTypeID)
}

func (ts *ContentfulManagementTestServer) SetEditorInterface(spaceID, environmentID, contentTypeID string, editorInterfaceFields cm.EditorInterfaceFields) {
	editorInterface := NewEditorInterfaceFromFields(spaceID, environmentID, contentTypeID, editorInterfaceFields)
	ts.editorInterfaces.Set(spaceID, environmentID, contentTypeID, &editorInterface)
}

func (ts *ContentfulManagementTestServer) DeleteEditorInterface(spaceID, environmentID, contentTypeID string) {
	ts.editorInterfaces.Delete(spaceID, environmentID, contentTypeID)
}
