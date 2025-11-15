package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetAppSigningSecret(organizationID, appDefinitionID string, fields cm.AppSigningSecretRequestData) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	appSigningSecret := NewAppSigningSecretFromRequest(organizationID, appDefinitionID, fields)
	s.h.appSigningSecrets.Set(organizationID, appDefinitionID, &appSigningSecret)
}
