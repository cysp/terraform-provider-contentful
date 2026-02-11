package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetContentType(spaceID, environmentID, contentTypeID string, contentTypeFields cm.ContentTypeRequestData) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	s.h.registerSpaceEnvironment(spaceID, environmentID, "ready")

	contentType := NewContentTypeFromRequestFields(spaceID, environmentID, contentTypeID, contentTypeFields)
	s.h.contentTypes.Set(spaceID, environmentID, contentTypeID, &contentType)
}
