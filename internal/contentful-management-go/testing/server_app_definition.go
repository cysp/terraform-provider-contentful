package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetAppDefinition(organizationID, appDefinitionID string, fields cm.AppDefinitionData) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	appDefinition := NewAppDefinitionFromFields(organizationID, appDefinitionID, fields)
	s.h.appDefinitions.Set(organizationID, appDefinitionID, &appDefinition)
}
