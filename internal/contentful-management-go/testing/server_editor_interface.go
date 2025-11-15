package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetEditorInterface(spaceID, environmentID, contentTypeID string, editorInterfaceFields cm.EditorInterfaceData) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	editorInterface := NewEditorInterfaceFromFields(spaceID, environmentID, contentTypeID, editorInterfaceFields)
	s.h.editorInterfaces.Set(spaceID, environmentID, contentTypeID, &editorInterface)
}
