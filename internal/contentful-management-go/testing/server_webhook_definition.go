package cmtesting

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetWebhookDefinition(spaceID string, webhookID string, fields cm.WebhookDefinitionData) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	webhook := NewWebhookDefinitionFromFields(spaceID, webhookID, fields)
	s.h.webhookDefinitions.Set(spaceID, webhookID, &webhook)
}

// StoredWebhookDefinition returns the mock server's unredacted representation.
// CMA responses still redact secret values.
func (s *Server) StoredWebhookDefinition(spaceID, webhookID string) (cm.WebhookDefinition, bool) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	stored := s.h.webhookDefinitions.Get(spaceID, webhookID)
	if stored == nil {
		return cm.WebhookDefinition{}, false
	}

	result := *stored
	result.Headers = append(cm.WebhookDefinitionHeaders(nil), stored.Headers...)

	return result, true
}
