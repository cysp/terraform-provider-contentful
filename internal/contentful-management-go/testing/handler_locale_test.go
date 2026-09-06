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

func TestCreateLocaleUsesGeneratedIdentityAndInitialVersion(t *testing.T) {
	t.Parallel()

	handler := cmt.NewHandler()
	handler.RegisterSpaceEnvironment("space", "environment", "ready")

	response, err := handler.CreateLocale(context.Background(), &cm.LocaleData{
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
	assert.Equal(t, 1, created.Response.Sys.Version)
}

func TestCreateLocaleRejectsDuplicateCode(t *testing.T) {
	t.Parallel()

	handler := cmt.NewHandler()
	handler.RegisterSpaceEnvironment("space", "environment", "ready")

	request := cm.LocaleData{
		Name:                 "English (United States)",
		Code:                 "en-US",
		FallbackCode:         cm.NewNilStringNull(),
		ContentDeliveryApi:   true,
		ContentManagementApi: true,
	}
	params := cm.CreateLocaleParams{
		SpaceID:       "space",
		EnvironmentID: "environment",
	}

	_, err := handler.CreateLocale(context.Background(), &request, params)
	require.NoError(t, err)

	response, err := handler.CreateLocale(context.Background(), &request, params)
	require.NoError(t, err)

	conflict, ok := response.(*cm.ErrorStatusCode)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnprocessableEntity, conflict.StatusCode)
}

func TestGetLocalesHonorsRequestedOrder(t *testing.T) {
	t.Parallel()

	handler := cmt.NewHandler()
	handler.RegisterSpaceEnvironment("space", "environment", "ready")

	for _, localeCode := range []string{"en-AU", "en-CA"} {
		_, err := handler.CreateLocale(context.Background(), &cm.LocaleData{
			Name:                 localeCode,
			Code:                 localeCode,
			FallbackCode:         cm.NewNilStringNull(),
			ContentDeliveryApi:   true,
			ContentManagementApi: true,
		}, cm.CreateLocaleParams{
			SpaceID:       "space",
			EnvironmentID: "environment",
		})
		require.NoError(t, err)
	}

	response, err := handler.GetLocales(context.Background(), cm.GetLocalesParams{
		SpaceID:       "space",
		EnvironmentID: "environment",
		Order:         []string{"-code"},
	})
	require.NoError(t, err)

	collection, ok := response.(*cm.LocaleCollection)
	require.True(t, ok)
	require.Len(t, collection.Items, 2)
	assert.Equal(t, []string{"en-CA", "en-AU"}, []string{collection.Items[0].Code, collection.Items[1].Code})
}
