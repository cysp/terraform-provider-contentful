package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetResourceProvider(organizationID, appDefinitionID string, fields cm.ResourceProviderRequest) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	appDefinition := NewResourceProviderFromRequest(organizationID, appDefinitionID, fields)
	s.h.appDefinitionResourceProviders[appDefinitionID] = &appDefinition
}
