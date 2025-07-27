package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetAppDefinitionResourceProvider(organizationID, appDefinitionID string, fields cm.ResourceProviderRequest) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	appDefinition := NewAppDefinitionResourceProviderFromRequest(organizationID, appDefinitionID, fields)
	s.h.appDefinitionResourceProviders.Set(organizationID, appDefinitionID, &appDefinition)
}
