package cmtesting_test

import (
	"context"
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeliveryAPIKeyVersionAndConflictSemantics(t *testing.T) {
	t.Parallel()

	handler := cmt.NewHandler()
	handler.RegisterSpaceEnvironment("space", "master", "ready")

	createdResponse, err := handler.CreateDeliveryAPIKey(context.Background(), &cm.ApiKeyRequestData{Name: "key"}, cm.CreateDeliveryAPIKeyParams{
		SpaceID: "space",
	})
	require.NoError(t, err)

	createdStatus, ok := createdResponse.(*cm.ApiKeyStatusCode)
	require.True(t, ok)
	assert.Equal(t, http.StatusCreated, createdStatus.StatusCode)
	assert.Equal(t, 0, createdStatus.Response.Sys.Version)

	updatedResponse, err := handler.UpdateDeliveryAPIKey(context.Background(), &cm.ApiKeyRequestData{Name: "updated"}, cm.UpdateDeliveryAPIKeyParams{
		SpaceID:            "space",
		APIKeyID:           createdStatus.Response.Sys.ID,
		XContentfulVersion: createdStatus.Response.Sys.Version,
	})
	require.NoError(t, err)

	updatedStatus, ok := updatedResponse.(*cm.ApiKeyStatusCode)
	require.True(t, ok)
	assert.Equal(t, http.StatusOK, updatedStatus.StatusCode)
	assert.Equal(t, 1, updatedStatus.Response.Sys.Version)

	staleResponse, err := handler.UpdateDeliveryAPIKey(context.Background(), &cm.ApiKeyRequestData{Name: "stale"}, cm.UpdateDeliveryAPIKeyParams{
		SpaceID:            "space",
		APIKeyID:           createdStatus.Response.Sys.ID,
		XContentfulVersion: createdStatus.Response.Sys.Version,
	})
	require.NoError(t, err)
	requireContentfulError(t, staleResponse, http.StatusConflict, cm.ErrorSysIDConflict, "Conflict")
}
