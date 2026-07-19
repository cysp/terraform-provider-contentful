package cmtesting

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetAppKey(organizationID, appDefinitionID string, request cm.AppKeyRequestData) error {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	appKey, err := NewAppKeyFromRequest(organizationID, appDefinitionID, request)
	if err != nil {
		return err
	}

	appKey.Generated.Reset()

	s.h.appKeys.Set(organizationID, appDefinitionID, &appKey)

	return nil
}
