package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetSpaceEnablements(spaceID string, spaceEnablementFields cm.SpaceEnablementData) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	spaceEnablements := NewSpaceEnablementFromRequestFields(spaceID, spaceEnablementFields)
	s.h.enablements[spaceID] = &spaceEnablements
}
