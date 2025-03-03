package contentfulmanagementtestserver

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) setupSpaceEnvironmentContentTypeHandlers() {
	ts.serveMux.Handle("/spaces/{spaceID}/environments/{environmentID}/content_types/{contentTypeID}", http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")
		environmentID := r.PathValue("environmentID")
		contentTypeID := r.PathValue("contentTypeID")

		switch r.Method {
		case http.MethodGet:
			contentType, exists := ts.contentTypes.Get(spaceID, environmentID, contentTypeID)
			if !exists {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, contentType)
		case http.MethodPut:
			var contentType cm.ContentType
			_ = ReadContentfulManagementRequest(r, &contentType)

			contentType.Sys = cm.ContentTypeSys{
				Type: cm.ContentTypeSysTypeContentType,
				ID:   contentTypeID,
			}

			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, &contentType)

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

		switch r.Method {
		case http.MethodPut:
			contentType, exists := ts.contentTypes.Get(spaceID, environmentID, contentTypeID)
			if !exists {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, contentType)

		case http.MethodDelete:
			responseWriter.WriteHeader(http.StatusNoContent)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)
		}
	}))

	ts.serveMux.Handle("/spaces/{spaceID}/environments/{environmentID}/content_types/{contentTypeID}/editor_interface", http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")
		environmentID := r.PathValue("environmentID")
		contentTypeID := r.PathValue("contentTypeID")

		switch r.Method {
		case http.MethodGet:
			editorInterface, exists := ts.editorInterfaces.Get(spaceID, environmentID, contentTypeID)
			if !exists {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, editorInterface)

		case http.MethodPut:
			editorInterface := cm.EditorInterface{}
			_ = ReadContentfulManagementRequest(r, &editorInterface)

			editorInterface.Sys = cm.EditorInterfaceSys{
				Type: cm.EditorInterfaceSysTypeEditorInterface,
				ID:   contentTypeID,
			}

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

func (ts *ContentfulManagementTestServer) GetEditorInterface(spaceID, environmentID, contentTypeID string) (*cm.EditorInterface, bool) {
	return ts.editorInterfaces.Get(spaceID, environmentID, contentTypeID)
}

func (ts *ContentfulManagementTestServer) SetContentType(spaceID, environmentID string, contentType *cm.ContentType) {
	ts.contentTypes.Set(spaceID, environmentID, contentType.Sys.ID, contentType)
}

func (ts *ContentfulManagementTestServer) SetEditorInterface(spaceID, environmentID, contentTypeID string, editorInterface *cm.EditorInterface) {
	ts.editorInterfaces.Set(spaceID, environmentID, contentTypeID, editorInterface)
}

func (ts *ContentfulManagementTestServer) DeleteContentType(spaceID, environmentID, contentTypeID string) {
	ts.contentTypes.Delete(spaceID, environmentID, contentTypeID)
	ts.DeleteEditorInterface(spaceID, environmentID, contentTypeID)
}

func (ts *ContentfulManagementTestServer) DeleteEditorInterface(spaceID, environmentID, contentTypeID string) {
	ts.editorInterfaces.Delete(spaceID, environmentID, contentTypeID)
}
