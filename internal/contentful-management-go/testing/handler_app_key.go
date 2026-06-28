package cmtesting

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetAppKeys(_ context.Context, params cm.GetAppKeysParams) (cm.GetAppKeysRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	keysByID := ts.appDefinitionKeys[params.AppDefinitionID]

	items := make([]cm.AppKey, 0, len(keysByID))
	for _, key := range keysByID {
		keyWithoutPrivateKey := *key
		keyWithoutPrivateKey.PrivateKey.Reset()
		items = append(items, keyWithoutPrivateKey)
	}

	return &cm.AppKeyCollection{
		Sys: cm.AppKeyCollectionSys{
			Type: cm.AppKeyCollectionSysTypeArray,
		},
		Total: len(items),
		Items: items,
	}, nil
}

//nolint:ireturn
func (ts *Handler) GetAppKey(_ context.Context, params cm.GetAppKeyParams) (cm.GetAppKeyRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	appKey := ts.appDefinitionKeys[params.AppDefinitionID][params.KeyKid]
	if appKey == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("AppKey not found"), nil), nil
	}

	keyWithoutPrivateKey := *appKey
	keyWithoutPrivateKey.PrivateKey.Reset()

	return &keyWithoutPrivateKey, nil
}

//nolint:ireturn
func (ts *Handler) CreateAppKey(_ context.Context, req *cm.AppKeyRequestData, params cm.CreateAppKeyParams) (cm.CreateAppKeyRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	appDefinition := ts.appDefinitions[params.AppDefinitionID]
	if appDefinition == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("AppDefinition not found"), nil), nil
	}

	appKey := NewAppKeyFromRequest(params.OrganizationID, params.AppDefinitionID, *req)

	if ts.appDefinitionKeys[params.AppDefinitionID] == nil {
		ts.appDefinitionKeys[params.AppDefinitionID] = make(map[string]*cm.AppKey)
	}

	ts.appDefinitionKeys[params.AppDefinitionID][appKey.Sys.ID] = &appKey

	return &appKey, nil
}

//nolint:ireturn
func (ts *Handler) DeleteAppKey(_ context.Context, params cm.DeleteAppKeyParams) (cm.DeleteAppKeyRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	appKey := ts.appDefinitionKeys[params.AppDefinitionID][params.KeyKid]
	if appKey == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("AppKey not found"), nil), nil
	}

	delete(ts.appDefinitionKeys[params.AppDefinitionID], params.KeyKid)

	return &cm.NoContent{}, nil
}
