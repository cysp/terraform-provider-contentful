package cmtesting

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetAppKey(organizationID, appDefinitionID string, request cm.AppKeyRequestData) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	appKey := NewAppKeyFromRequest(organizationID, appDefinitionID, request)

	if s.h.appDefinitionKeys[appDefinitionID] == nil {
		s.h.appDefinitionKeys[appDefinitionID] = make(map[string]*cm.AppKey)
	}

	s.h.appDefinitionKeys[appDefinitionID][appKey.Sys.ID] = &appKey
}
