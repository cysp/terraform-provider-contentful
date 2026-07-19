package cmtesting

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetAppKey(organizationID, appDefinitionID string, request cm.AppKeyRequestData) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	appKey := NewAppKeyFromRequest(organizationID, appDefinitionID, request)

	s.h.appKeys.Set(organizationID, appDefinitionID, appKey)
}
