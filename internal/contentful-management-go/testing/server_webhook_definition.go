package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetWebhookDefinition(spaceID string, webhookID string, fields cm.WebhookDefinitionFields) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	webhook := NewWebhookDefinitionFromFields(spaceID, webhookID, fields)
	s.h.webhookDefinitions.Set(spaceID, webhookID, &webhook)
}
