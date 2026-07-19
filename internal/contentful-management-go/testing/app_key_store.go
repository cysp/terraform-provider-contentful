package cmtesting

import (
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

type appKeyStore struct {
	records []appKeyRecord
}

type appKeyRecord struct {
	organizationID  string
	appDefinitionID string
	appKey          cm.AppKey
}

func (s *appKeyStore) Contains(keyKID string) bool {
	return slices.ContainsFunc(s.records, func(record appKeyRecord) bool {
		return record.appKey.Sys.ID == keyKID
	})
}

func (s *appKeyStore) Get(organizationID, appDefinitionID, keyKID string) (cm.AppKey, bool) {
	index := s.index(organizationID, appDefinitionID, keyKID)
	if index == -1 {
		return cm.AppKey{}, false
	}

	return s.records[index].appKey, true
}

func (s *appKeyStore) Set(organizationID, appDefinitionID string, appKey cm.AppKey) {
	record := appKeyRecord{
		organizationID:  organizationID,
		appDefinitionID: appDefinitionID,
		appKey:          appKey,
	}

	index := slices.IndexFunc(s.records, func(existing appKeyRecord) bool {
		return existing.appKey.Sys.ID == appKey.Sys.ID
	})
	if index == -1 {
		s.records = append(s.records, record)

		return
	}

	s.records[index] = record
}

func (s *appKeyStore) Delete(organizationID, appDefinitionID, keyKID string) {
	index := s.index(organizationID, appDefinitionID, keyKID)
	if index != -1 {
		s.records = slices.Delete(s.records, index, index+1)
	}
}

func (s *appKeyStore) List(organizationID, appDefinitionID string) []cm.AppKey {
	appKeys := make([]cm.AppKey, 0, len(s.records))

	for _, record := range s.records {
		if record.organizationID == organizationID && record.appDefinitionID == appDefinitionID {
			appKeys = append(appKeys, record.appKey)
		}
	}

	return appKeys
}

func (s *appKeyStore) index(organizationID, appDefinitionID, keyKID string) int {
	return slices.IndexFunc(s.records, func(record appKeyRecord) bool {
		return record.organizationID == organizationID &&
			record.appDefinitionID == appDefinitionID &&
			record.appKey.Sys.ID == keyKID
	})
}
