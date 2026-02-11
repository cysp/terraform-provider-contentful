package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetEnvironment(spaceID, environmentID, status string, data cm.EnvironmentData) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	environment := NewEnvironmentFromEnvironmentData(spaceID, environmentID, status, data)
	s.h.environments.Set(spaceID, environmentID, &environment)
}
