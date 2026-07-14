package cmtesting

import (
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

type previewEnvironmentStore struct {
	values cm.SpaceMap[*cm.PreviewEnvironment]
	ids    map[string][]string
}

func newPreviewEnvironmentStore() previewEnvironmentStore {
	return previewEnvironmentStore{
		values: cm.NewSpaceMap[*cm.PreviewEnvironment](),
		ids:    make(map[string][]string),
	}
}

func (s *previewEnvironmentStore) Get(spaceID, previewEnvironmentID string) *cm.PreviewEnvironment {
	return s.values.Get(spaceID, previewEnvironmentID)
}

func (s *previewEnvironmentStore) Set(spaceID, previewEnvironmentID string, previewEnvironment *cm.PreviewEnvironment) {
	if s.Get(spaceID, previewEnvironmentID) == nil {
		s.ids[spaceID] = append(s.ids[spaceID], previewEnvironmentID)
	}

	s.values.Set(spaceID, previewEnvironmentID, previewEnvironment)
}

func (s *previewEnvironmentStore) Delete(spaceID, previewEnvironmentID string) {
	s.values.Delete(spaceID, previewEnvironmentID)

	if index := slices.Index(s.ids[spaceID], previewEnvironmentID); index >= 0 {
		s.ids[spaceID] = slices.Delete(s.ids[spaceID], index, index+1)
	}
}

func (s *previewEnvironmentStore) IDs(spaceID string) []string {
	return s.ids[spaceID]
}
