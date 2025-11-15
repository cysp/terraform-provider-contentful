package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetTeamSpaceMembership(spaceID, teamSpaceMembershipID, teamID string, fields cm.TeamSpaceMembershipData) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	teamSpaceMembership := NewTeamSpaceMembershipFromFields(spaceID, teamSpaceMembershipID, teamID, fields)
	s.h.teamSpaceMemberships.Set(spaceID, teamSpaceMembershipID, &teamSpaceMembership)
}
