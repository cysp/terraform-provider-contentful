package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (s *Server) SetEntry(spaceID, environmentID, contentTypeID, entryID string, req cm.EntryRequest) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	s.h.registerSpaceEnvironment(spaceID, environmentID)

	entry := NewEntryFromRequest(spaceID, environmentID, contentTypeID, entryID, &req)
	s.h.entries.Set(spaceID, environmentID, entryID, &entry)
}
