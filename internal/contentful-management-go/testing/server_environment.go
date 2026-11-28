package cmtesting

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetEnvironment(spaceID, environmentID, status string, data cm.EnvironmentData) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	environment := NewEnvironmentFromEnvironmentData(spaceID, environmentID, status, data)
	s.h.environments.Set(spaceID, environmentID, &environment)
}

// RegisterMasterEnvironmentAlias seeds master and its concrete target in an
// alias-enabled Space fixture. Use it in a Space without a concrete master,
// before seeding its children. Seed those children under the returned target ID.
func (s *Server) RegisterMasterEnvironmentAlias(spaceID string) string {
	const targetID = "master-1970-01-01"

	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	s.h.registerSpaceEnvironment(spaceID, targetID, "ready")
	alias := NewEnvironmentAliasFromEnvironmentAliasData(spaceID, "master", cm.EnvironmentAliasData{Environment: cm.NewEnvironmentLink(targetID)})
	s.h.environmentAliases.Set(spaceID, "master", &alias)

	return targetID
}
