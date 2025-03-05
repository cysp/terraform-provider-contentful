package contentfulmanagementtestserver

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

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

		contentType, exists := ts.contentTypes.Get(spaceID, environmentID, contentTypeID)

		switch r.Method {
		case http.MethodGet:
			if !exists {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, contentType)
		case http.MethodPut:
			var contentTypeRequestFields cm.ContentTypeRequestFields
			if err := ReadContentfulManagementRequest(r, &contentTypeRequestFields); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(responseWriter, err)

				return
			}

			if exists {
				UpdateContentTypeFromRequestFields(contentType, contentTypeRequestFields)

				_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, contentType)
			} else {
				contentType := NewContentTypeFromRequestFields(spaceID, environmentID, contentTypeID, contentTypeRequestFields)

				ts.contentTypes.Set(spaceID, environmentID, contentTypeID, &contentType)

				_ = WriteContentfulManagementResponse(responseWriter, http.StatusCreated, &contentType)
			}

		case http.MethodDelete:
			if !exists {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

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

		contentType, exists := ts.contentTypes.Get(spaceID, environmentID, contentTypeID)

		switch r.Method {
		case http.MethodPut:
			if !exists {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			contentType.Sys.PublishedVersion.SetTo(contentType.Sys.Version)

			contentType.Sys.Version++

			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, contentType)

		case http.MethodDelete:
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
