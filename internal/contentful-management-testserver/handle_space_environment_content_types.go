package contentfulmanagementtestserver

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:cyclop,gocognit
func (ts *ContentfulManagementTestServer) setupSpaceEnvironmentContentTypeHandlers() {
	ts.serveMux.Handle("/spaces/{spaceID}/environments/{environmentID}/content_types/{contentTypeID}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")
		environmentID := r.PathValue("environmentID")
		contentTypeID := r.PathValue("contentTypeID")

		if spaceID == NonexistentID || environmentID == NonexistentID || contentTypeID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(w)

			return
		}

		ts.mu.RLock()
		defer ts.mu.RUnlock()

		contentType := ts.contentTypes.Get(spaceID, environmentID, contentTypeID)

		switch r.Method {
		case http.MethodGet:
			switch contentType {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				_ = WriteContentfulManagementResponse(w, http.StatusOK, contentType)
			}
		case http.MethodPut:
			var contentTypeRequestFields cm.ContentTypeRequestFields
			if err := ReadContentfulManagementRequest(r, &contentTypeRequestFields); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(w, err)

				return
			}

			switch contentType {
			case nil:
				contentType := NewContentTypeFromRequestFields(spaceID, environmentID, contentTypeID, contentTypeRequestFields)
				ts.contentTypes.Set(spaceID, environmentID, contentTypeID, &contentType)
				_ = WriteContentfulManagementResponse(w, http.StatusCreated, &contentType)
			default:
				UpdateContentTypeFromRequestFields(contentType, contentTypeRequestFields)
				_ = WriteContentfulManagementResponse(w, http.StatusOK, contentType)
			}

		case http.MethodDelete:
			switch contentType {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				ts.contentTypes.Delete(spaceID, environmentID, contentTypeID)
				w.WriteHeader(http.StatusNoContent)
			}

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))

	ts.serveMux.Handle("/spaces/{spaceID}/environments/{environmentID}/content_types/{contentTypeID}/published", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")
		environmentID := r.PathValue("environmentID")
		contentTypeID := r.PathValue("contentTypeID")

		if spaceID == NonexistentID || environmentID == NonexistentID || contentTypeID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(w)

			return
		}

		ts.mu.RLock()
		defer ts.mu.RUnlock()

		contentType := ts.contentTypes.Get(spaceID, environmentID, contentTypeID)

		switch r.Method {
		case http.MethodPut:
			switch contentType {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				publishContentType(contentType)
				_ = WriteContentfulManagementResponse(w, http.StatusOK, contentType)
			}

		case http.MethodDelete:
			switch contentType {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				contentType.Sys.PublishedVersion.Reset()
				w.WriteHeader(http.StatusNoContent)
			}

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))

	ts.serveMux.Handle("/spaces/{spaceID}/environments/{environmentID}/content_types/{contentTypeID}/editor_interface", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")
		environmentID := r.PathValue("environmentID")
		contentTypeID := r.PathValue("contentTypeID")

		if spaceID == NonexistentID || environmentID == NonexistentID || contentTypeID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(w)

			return
		}

		ts.mu.RLock()
		defer ts.mu.RUnlock()

		contentType := ts.contentTypes.Get(spaceID, environmentID, contentTypeID)
		editorInterface := ts.editorInterfaces.Get(spaceID, environmentID, contentTypeID)

		switch r.Method {
		case http.MethodGet:
			switch editorInterface {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				_ = WriteContentfulManagementResponse(w, http.StatusOK, editorInterface)
			}

		case http.MethodPut:
			if contentType == nil {
				_ = WriteContentfulManagementErrorNotFoundResponse(w)

				return
			}

			editorInterfaceFields := cm.EditorInterfaceFields{}
			if err := ReadContentfulManagementRequest(r, &editorInterfaceFields); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(w, err)

				return
			}

			switch editorInterface {
			case nil:
				editorInterface := NewEditorInterfaceFromFields(spaceID, environmentID, contentTypeID, editorInterfaceFields)
				ts.editorInterfaces.Set(spaceID, environmentID, contentTypeID, &editorInterface)
				_ = WriteContentfulManagementResponse(w, http.StatusOK, &editorInterface)
			default:
				UpdateEditorInterfaceFromFields(editorInterface, editorInterfaceFields)
				_ = WriteContentfulManagementResponse(w, http.StatusOK, editorInterface)
			}

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))
}

func (ts *ContentfulManagementTestServer) SetContentType(spaceID, environmentID, contentTypeID string, contentTypeFields cm.ContentTypeRequestFields) {
	contentType := NewContentTypeFromRequestFields(spaceID, environmentID, contentTypeID, contentTypeFields)
	ts.contentTypes.Set(spaceID, environmentID, contentType.Sys.ID, &contentType)
}

func (ts *ContentfulManagementTestServer) SetEditorInterface(spaceID, environmentID, contentTypeID string, editorInterfaceFields cm.EditorInterfaceFields) {
	editorInterface := NewEditorInterfaceFromFields(spaceID, environmentID, contentTypeID, editorInterfaceFields)
	ts.editorInterfaces.Set(spaceID, environmentID, contentTypeID, &editorInterface)
}
