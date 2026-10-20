package cmtesting

import (
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetAppInstallation(spaceID, environmentID, appDefinitionID string, fields cm.AppInstallationData) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	installation := NewAppInstallationFromFields(spaceID, environmentID, appDefinitionID, fields)
	s.h.appInstallations.Set(spaceID, environmentID, appDefinitionID, &installation)
}

// StoredAppInstallation returns a copy of the mock server's unredacted representation.
// CMA responses still redact Secret values according to the current App Definition.
func (s *Server) StoredAppInstallation(spaceID, environmentID, appDefinitionID string) (cm.AppInstallation, bool) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	stored := s.h.appInstallations.Get(spaceID, environmentID, appDefinitionID)
	if stored == nil {
		return cm.AppInstallation{}, false
	}

	result := *stored
	result.Parameters = slices.Clone(stored.Parameters)

	return result, true
}
