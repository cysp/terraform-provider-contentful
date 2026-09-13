package cmtesting_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest,tparallel // These steps intentionally share and mutate one parent lifecycle.
func TestAppEventSubscriptionParentAndValidation(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	handler := server.Handler()
	params := cm.PutAppEventSubscriptionParams{OrganizationID: "organization", AppDefinitionID: "app"}
	request := cm.AppEventSubscriptionData{Topics: []string{"Entry.publish"}, TargetUrl: cm.NewOptString("https://example.invalid/events")}
	result, err := handler.PutAppEventSubscription(t.Context(), &request, params)
	require.NoError(t, err)
	appEventAssertNotFound(t, result)
	server.SetAppDefinition("organization", "app", cm.AppDefinitionData{Name: "App"})

	result, err = handler.PutAppEventSubscription(t.Context(), &request, params)
	require.NoError(t, err)
	require.IsType(t, &cm.PutAppEventSubscriptionCreated{}, result)

	for name, invalid := range map[string]cm.AppEventSubscriptionData{
		"no topics":           {TargetUrl: request.TargetUrl},
		"duplicate topics":    {Topics: []string{"Entry.publish", "Entry.publish"}, TargetUrl: request.TargetUrl},
		"empty topic":         {Topics: []string{""}, TargetUrl: request.TargetUrl},
		"missing HTTP target": {Topics: request.Topics},
		"empty target":        {Topics: request.Topics, TargetUrl: cm.NewOptString("")},
		"non HTTPS target":    {Topics: request.Topics, TargetUrl: cm.NewOptString("http://example.invalid/events")},
	} {
		t.Run(name, func(t *testing.T) {
			result, err := handler.PutAppEventSubscription(t.Context(), &invalid, params)
			require.NoError(t, err)
			require.IsType(t, &cm.ErrorStatusCode{}, result)
			resultRead, err := handler.GetAppEventSubscription(t.Context(), cm.GetAppEventSubscriptionParams(params))
			require.NoError(t, err)

			actual, ok := resultRead.(*cm.AppEventSubscription)
			require.True(t, ok)
			assert.Equal(t, []string{"Entry.publish"}, actual.Topics)
			assert.Equal(t, "https://example.invalid/events", actual.TargetUrl.Value)
		})
	}

	wrong := cm.PutAppEventSubscriptionParams{OrganizationID: "other", AppDefinitionID: "app"}
	result, err = handler.PutAppEventSubscription(t.Context(), &request, wrong)
	require.NoError(t, err)
	appEventAssertNotFound(t, result)
	read, err := handler.GetAppEventSubscription(t.Context(), cm.GetAppEventSubscriptionParams(wrong))
	require.NoError(t, err)
	appEventAssertNotFound(t, read)
	deleted, err := handler.DeleteAppEventSubscription(t.Context(), cm.DeleteAppEventSubscriptionParams(wrong))
	require.NoError(t, err)
	appEventAssertNotFound(t, deleted)
	_, err = handler.DeleteAppDefinition(t.Context(), cm.DeleteAppDefinitionParams{OrganizationID: "organization", AppDefinitionID: "app"})
	require.NoError(t, err)
	read, err = handler.GetAppEventSubscription(t.Context(), cm.GetAppEventSubscriptionParams(params))
	require.NoError(t, err)
	appEventAssertNotFound(t, read)
	deleted, err = handler.DeleteAppEventSubscription(t.Context(), cm.DeleteAppEventSubscriptionParams(params))
	require.NoError(t, err)
	appEventAssertNotFound(t, deleted)
}

func appEventAssertNotFound(t *testing.T, result any) {
	t.Helper()

	status, ok := result.(cm.StatusCodeResponse)
	require.True(t, ok)
	assert.Equal(t, 404, status.GetStatusCode())
}
