package cmtesting_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/require"
)

func requireAppDefinitionDoesNotExistDetails(t *testing.T, response any) {
	t.Helper()

	errorStatus, ok := response.(*cm.ErrorStatusCode)
	require.True(t, ok)
	errorResponse, ok := errorStatus.Response.GetError()
	require.True(t, ok)
	require.Contains(t, string(errorResponse.Details), "AppDefinition does not exist.")
}

func newAppKeyTestHandler(t *testing.T) *cmt.Handler {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)
	server.SetAppDefinition("organization", "app-definition", cm.AppDefinitionData{Name: "App"})

	return server.Handler()
}

func appKeyCreateParams() cm.CreateAppKeyParams {
	return cm.CreateAppKeyParams{
		OrganizationID:  "organization",
		AppDefinitionID: "app-definition",
	}
}

func createAppKey(t *testing.T, handler *cmt.Handler, request *cm.AppKeyRequestData) cm.AppKey {
	t.Helper()

	response, err := handler.CreateAppKey(t.Context(), request, appKeyCreateParams())
	require.NoError(t, err)

	appKey, ok := response.(*cm.AppKey)
	require.True(t, ok)

	return *appKey
}
