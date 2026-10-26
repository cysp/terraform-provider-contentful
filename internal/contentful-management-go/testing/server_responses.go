package cmtesting

import cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"

// SetContentTypeResponse stores an exact response fixture without request validation
// or lifecycle transitions. It owns a detached copy, including adversarial values.
func (s *Server) SetContentTypeResponse(response cm.ContentType) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	snapshot := snapshotResponse(response)
	s.h.contentTypes.Set(response.Sys.Space.Sys.ID, response.Sys.Environment.Sys.ID, response.Sys.ID, &snapshot)
}

// SetRoleResponse stores an exact response fixture, including its version, without
// request handling. It owns a detached copy of the supplied response.
func (s *Server) SetRoleResponse(response cm.Role) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	snapshot := snapshotResponse(response)
	s.h.roles.Set(response.Sys.Space.Sys.ID, response.Sys.ID, &snapshot)
}

// SetTaxonomyConceptResponse stores an exact response fixture without request
// normalization or membership updates. It owns a detached copy of the response.
func (s *Server) SetTaxonomyConceptResponse(response cm.TaxonomyConcept) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	snapshot := snapshotResponse(response)
	s.h.taxonomyConcepts.Set(response.Sys.Organization.Sys.ID, response.Sys.ID, &snapshot)
}

// SetTaxonomyConceptSchemeResponse stores an exact response fixture without
// request normalization or membership updates. It owns a detached response copy.
func (s *Server) SetTaxonomyConceptSchemeResponse(response cm.TaxonomyConceptScheme) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	snapshot := snapshotResponse(response)
	s.h.taxonomyConceptSchemes.Set(response.Sys.Organization.Sys.ID, response.Sys.ID, &snapshot)
}
