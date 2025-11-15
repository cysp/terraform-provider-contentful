package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetRole(spaceID, roleID string, fields cm.RoleData) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	role := NewRoleFromFields(spaceID, roleID, fields)
	s.h.roles.Set(spaceID, roleID, &role)
}
