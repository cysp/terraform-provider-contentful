package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetMarketplaceAppDefinition(organizationID, appDefinitionID string, fields cm.AppDefinitionFields) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	appDefinition := NewAppDefinitionFromFields(organizationID, appDefinitionID, fields)
	s.h.marketplaceAppDefinitions[appDefinitionID] = &appDefinition
}
