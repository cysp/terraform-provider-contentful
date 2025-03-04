package contentfulmanagementtestserver

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:cyclop,gocognit,maintidx
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
			if environmentID == NonexistentID {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			_, exists := ts.contentTypes.Get(spaceID, environmentID, contentTypeID)

			var contentTypeRequestFields cm.ContentTypeRequestFields
			if err := ReadContentfulManagementRequest(r, &contentTypeRequestFields); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(responseWriter, err)

				return
			}

			contentTypeFields := make([]cm.ContentTypeFieldsItem, len(contentTypeRequestFields.Fields))

			for fieldIndex, field := range contentTypeRequestFields.Fields {
				contentTypeFieldItems := cm.OptContentTypeFieldsItemItems{}

				if fieldItems, fieldItemsOk := field.Items.Get(); fieldItemsOk {
					contentTypeFieldItemItems := cm.ContentTypeFieldsItemItems{}

					contentTypeFieldItemItems.Type = fieldItems.Type
					contentTypeFieldItemItems.LinkType = fieldItems.LinkType
					contentTypeFieldItemItems.Validations = fieldItems.Validations

					contentTypeFieldItems.SetTo(contentTypeFieldItemItems)
				}

				contentTypeFields[fieldIndex] = cm.ContentTypeFieldsItem{
					ID:           field.ID,
					Name:         field.Name,
					Type:         field.Type,
					LinkType:     field.LinkType,
					Items:        contentTypeFieldItems,
					Localized:    field.Localized,
					Required:     field.Required,
					Validations:  field.Validations,
					Omitted:      field.Omitted,
					Disabled:     field.Disabled,
					DefaultValue: field.DefaultValue,
				}
			}

			contentType := cm.ContentType{
				Sys: cm.ContentTypeSys{
					Type: cm.ContentTypeSysTypeContentType,
					ID:   contentTypeID,
				},
				Name:         contentTypeRequestFields.Name,
				Description:  contentTypeRequestFields.Description,
				Fields:       contentTypeFields,
				DisplayField: cm.NewNilString(contentTypeRequestFields.DisplayField),
			}

			ts.contentTypes.Set(spaceID, environmentID, contentTypeID, &contentType)

			statusCode := http.StatusCreated
			if exists {
				statusCode = http.StatusOK
			}

			_ = WriteContentfulManagementResponse(responseWriter, statusCode, &contentType)

		case http.MethodDelete:
			if environmentID == NonexistentID {
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

		switch r.Method {
		case http.MethodPut:
			if environmentID == NonexistentID {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

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
			if environmentID == NonexistentID {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			_, exists := ts.contentTypes.Get(spaceID, environmentID, contentTypeID)
			if !exists {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			editorInterfaceFields := cm.EditorInterfaceFields{}
			_ = ReadContentfulManagementRequest(r, &editorInterfaceFields)

			editorInterface := cm.EditorInterface{
				Sys: cm.EditorInterfaceSys{
					Type: cm.EditorInterfaceSysTypeEditorInterface,
					ID:   contentTypeID,
				},
				// EditorLayout:  editorInterfaceFields.EditorLayout,
				// Controls:      editorInterfaceFields.Controls,
				// GroupControls: editorInterfaceFields.GroupControls,
				// Sidebar:       editorInterfaceFields.Sidebar,
			}

			if editorLayout, ok := editorInterfaceFields.EditorLayout.Get(); ok {
				editorLayoutItemArray := make([]cm.EditorInterfaceEditorLayoutItem, len(editorLayout))

				for index, editorLayoutItem := range editorLayout {
					editorLayoutItemArray[index] = cm.EditorInterfaceEditorLayoutItem(editorLayoutItem)
				}

				editorInterface.EditorLayout = cm.NewOptNilEditorInterfaceEditorLayoutItemArray(editorLayoutItemArray)
			}

			if controls, ok := editorInterfaceFields.Controls.Get(); ok {
				controlsArray := make([]cm.EditorInterfaceControlsItem, len(controls))

				for index, control := range controls {
					controlsArray[index] = cm.EditorInterfaceControlsItem(control)
				}

				editorInterface.Controls = cm.NewOptNilEditorInterfaceControlsItemArray(controlsArray)
			}

			if groupControls, ok := editorInterfaceFields.GroupControls.Get(); ok {
				groupControlsArray := make([]cm.EditorInterfaceGroupControlsItem, len(groupControls))

				for index, groupControl := range groupControls {
					groupControlsArray[index] = cm.EditorInterfaceGroupControlsItem(groupControl)
				}

				editorInterface.GroupControls = cm.NewOptNilEditorInterfaceGroupControlsItemArray(groupControlsArray)
			}

			if sidebar, ok := editorInterfaceFields.Sidebar.Get(); ok {
				sidebarArray := make([]cm.EditorInterfaceSidebarItem, len(sidebar))

				for index, sidebarItem := range sidebar {
					sidebarArray[index] = cm.EditorInterfaceSidebarItem(sidebarItem)
				}

				editorInterface.Sidebar = cm.NewOptNilEditorInterfaceSidebarItemArray(sidebarArray)
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

func (ts *ContentfulManagementTestServer) SetEditorInterface(spaceID, environmentID string, editorInterface *cm.EditorInterface) {
	ts.editorInterfaces.Set(spaceID, environmentID, editorInterface.Sys.ID, editorInterface)
}

func (ts *ContentfulManagementTestServer) DeleteContentType(spaceID, environmentID, contentTypeID string) {
	ts.contentTypes.Delete(spaceID, environmentID, contentTypeID)
	ts.DeleteEditorInterface(spaceID, environmentID, contentTypeID)
}

func (ts *ContentfulManagementTestServer) DeleteEditorInterface(spaceID, environmentID, contentTypeID string) {
	ts.editorInterfaces.Delete(spaceID, environmentID, contentTypeID)
}
