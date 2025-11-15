package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetTeam(organizationID, teamID string, fields cm.TeamData) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	team := NewTeamFromFields(organizationID, teamID, fields)
	s.h.teams.Set(organizationID, teamID, &team)
}
