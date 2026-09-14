package cmtesting_test

import (
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateLocaleIdentityVersionAndDuplicateCode(t *testing.T) {
	t.Parallel()

	handler := cmt.NewHandler()
	handler.RegisterSpaceEnvironment("space", "environment", "ready")

	response, err := handler.CreateLocale(t.Context(), &cm.LocaleData{
		Name:                 "English (United States)",
		Code:                 "en-US",
		FallbackCode:         cm.NewNilStringNull(),
		ContentDeliveryApi:   true,
		ContentManagementApi: true,
		Optional:             false,
	}, cm.CreateLocaleParams{
		SpaceID:       "space",
		EnvironmentID: "environment",
	})
	require.NoError(t, err)

	created, ok := response.(*cm.LocaleStatusCode)
	require.True(t, ok)
	assert.NotEqual(t, created.Response.Code, created.Response.Sys.ID)
	assert.Equal(t, 1, created.Response.Sys.Version.Or(0))
	response, err = handler.CreateLocale(t.Context(), &cm.LocaleData{Code: "en-US"}, cm.CreateLocaleParams{SpaceID: "space", EnvironmentID: "environment"})
	require.NoError(t, err)

	conflict, ok := response.(*cm.ErrorStatusCode)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnprocessableEntity, conflict.StatusCode)
}
