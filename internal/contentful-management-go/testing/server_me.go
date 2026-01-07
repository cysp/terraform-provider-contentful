package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetMe(me *cm.User) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	s.h.me = me
}
