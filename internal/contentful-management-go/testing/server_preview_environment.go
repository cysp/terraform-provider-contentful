package cmtesting

import cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"

func (s *Server) SetPreviewEnvironment(spaceID, previewEnvironmentID string, data cm.PreviewEnvironmentData) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	previewEnvironment := NewPreviewEnvironmentFromData(spaceID, previewEnvironmentID, data)
	s.h.previewEnvironments.Set(spaceID, previewEnvironmentID, &previewEnvironment)
}

func (s *Server) DeletePreviewEnvironment(spaceID, previewEnvironmentID string) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	s.h.previewEnvironments.Delete(spaceID, previewEnvironmentID)
}

func (s *Server) IncrementPreviewEnvironmentVersion(spaceID, previewEnvironmentID string) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	if previewEnvironment := s.h.previewEnvironments.Get(spaceID, previewEnvironmentID); previewEnvironment != nil {
		previewEnvironment.Sys.Version++
	}
}
