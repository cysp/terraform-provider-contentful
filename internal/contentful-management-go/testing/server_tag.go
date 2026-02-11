package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetTag(spaceID, environmentID, tagID string, request cm.TagRequest) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	s.h.registerSpaceEnvironment(spaceID, environmentID, "ready")

	tag := NewTagFromRequest(spaceID, environmentID, tagID, &request)
	s.h.tags.Set(spaceID, environmentID, tagID, &tag)
}
