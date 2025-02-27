package testserver

import (
	"fmt"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) HandleContentType(spaceID string, environmentID string, contentType *cm.ContentType, editorInterface *cm.EditorInterface) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	contentTypeID := contentType.Sys.ID

	ts.ServeMux.Handle(fmt.Sprintf("/spaces/%s/environments/%s/content_types/%s", spaceID, environmentID, contentTypeID), http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, contentType)
		case http.MethodPut:
			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, contentType)
		case http.MethodDelete:
			responseWriter.WriteHeader(http.StatusNoContent)
		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)
		}
	}))

	ts.ServeMux.Handle(fmt.Sprintf("/spaces/%s/environments/%s/content_types/%s/published", spaceID, environmentID, contentTypeID), http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, contentType)
		case http.MethodDelete:
			responseWriter.WriteHeader(http.StatusNoContent)
		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)
		}
	}))

	ts.ServeMux.Handle(fmt.Sprintf("/spaces/%s/environments/%s/content_types/%s/editor_interface", spaceID, environmentID, contentTypeID), http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, editorInterface)
		case http.MethodPut:
			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, editorInterface)
		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)
		}
	}))
}
