package cmtesting

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

const appKeyMockUserID = "mock-user"

func NewAppKeyFromRequest(organizationID, appDefinitionID string, request cm.AppKeyRequestData) cm.AppKey {
	return cm.AppKey{
		Sys: cm.NewAppKeySys(organizationID, appDefinitionID, request.Jwk.Kid, appKeyMockUserID),
		Jwk: request.Jwk,
	}
}

type AppKeyMap struct {
	organizations map[string]map[string]*appKeyCollection
	ownersByID    map[string]appKeyOwner
}

type appKeyOwner struct {
	organizationID  string
	appDefinitionID string
}

type appKeyCollection struct {
	byID  map[string]*cm.AppKey
	order []string
}

func NewAppKeyMap() AppKeyMap {
	return AppKeyMap{
		organizations: make(map[string]map[string]*appKeyCollection),
		ownersByID:    make(map[string]appKeyOwner),
	}
}

func (m *AppKeyMap) Contains(keyKID string) bool {
	_, ok := m.ownersByID[keyKID]

	return ok
}

func (m *AppKeyMap) Get(organizationID, appDefinitionID, keyKID string) *cm.AppKey {
	collection := m.getCollection(organizationID, appDefinitionID)
	if collection == nil {
		return nil
	}

	return collection.byID[keyKID]
}

func (m *AppKeyMap) Set(organizationID, appDefinitionID string, appKey *cm.AppKey) {
	if m.organizations[organizationID] == nil {
		m.organizations[organizationID] = make(map[string]*appKeyCollection)
	}

	collection := m.organizations[organizationID][appDefinitionID]
	if collection == nil {
		collection = &appKeyCollection{byID: make(map[string]*cm.AppKey)}
		m.organizations[organizationID][appDefinitionID] = collection
	}

	if collection.byID[appKey.Sys.ID] == nil {
		collection.order = append(collection.order, appKey.Sys.ID)
	}

	collection.byID[appKey.Sys.ID] = appKey
	m.ownersByID[appKey.Sys.ID] = appKeyOwner{
		organizationID:  organizationID,
		appDefinitionID: appDefinitionID,
	}
}

func (m *AppKeyMap) Delete(organizationID, appDefinitionID, keyKID string) {
	collection := m.getCollection(organizationID, appDefinitionID)
	if collection == nil {
		return
	}

	delete(collection.byID, keyKID)

	owner, ok := m.ownersByID[keyKID]
	if ok && owner.organizationID == organizationID && owner.appDefinitionID == appDefinitionID {
		delete(m.ownersByID, keyKID)
	}

	for idx, id := range collection.order {
		if id == keyKID {
			collection.order = append(collection.order[:idx], collection.order[idx+1:]...)

			break
		}
	}
}

func (m *AppKeyMap) List(organizationID, appDefinitionID string) []cm.AppKey {
	collection := m.getCollection(organizationID, appDefinitionID)
	if collection == nil {
		return nil
	}

	// Preserve insertion order for deterministic mock pagination. Contentful does not document an ordering guarantee.
	appKeys := make([]cm.AppKey, 0, len(collection.byID))
	for _, id := range collection.order {
		appKeys = append(appKeys, *collection.byID[id])
	}

	return appKeys
}

func (m *AppKeyMap) getCollection(organizationID, appDefinitionID string) *appKeyCollection {
	organizationCollections, ok := m.organizations[organizationID]
	if !ok {
		return nil
	}

	return organizationCollections[appDefinitionID]
}
