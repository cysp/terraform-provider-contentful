//nolint:dupl
package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetAppSigningSecret(_ context.Context, params cm.GetAppSigningSecretParams) (cm.GetAppSigningSecretRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	appSigningSecret := ts.appSigningSecrets[params.AppDefinitionID]
	if appSigningSecret == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("AppSigningSecret not found"), nil), nil
	}

	return appSigningSecret, nil
}

//nolint:ireturn
func (ts *Handler) PutAppSigningSecret(_ context.Context, req *cm.AppSigningSecretRequestData, params cm.PutAppSigningSecretParams) (cm.PutAppSigningSecretRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	appDefinition := ts.appDefinitions[params.AppDefinitionID]
	if appDefinition == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("AppDefinition not found"), nil), nil
	}

	appSigningSecret := ts.appSigningSecrets[params.AppDefinitionID]
	if appSigningSecret == nil {
		appSigningSecret := NewAppSigningSecretFromRequest(params.OrganizationID, params.AppDefinitionID, *req)

		ts.appSigningSecrets[params.AppDefinitionID] = &appSigningSecret

		return &cm.AppSigningSecretStatusCode{
			StatusCode: http.StatusCreated,
			Response:   appSigningSecret,
		}, nil
	}

	UpdateAppSigningSecretFromRequest(appSigningSecret, params.OrganizationID, params.AppDefinitionID, *req)

	return &cm.AppSigningSecretStatusCode{
		StatusCode: http.StatusOK,
		Response:   *appSigningSecret,
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteAppSigningSecret(_ context.Context, params cm.DeleteAppSigningSecretParams) (cm.DeleteAppSigningSecretRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	appSigningSecret := ts.appSigningSecrets[params.AppDefinitionID]
	if appSigningSecret == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("AppSigningSecret not found"), nil), nil
	}

	delete(ts.appSigningSecrets, params.AppDefinitionID)

	return &cm.NoContent{}, nil
}
