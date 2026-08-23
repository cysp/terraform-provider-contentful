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

func TestGetTeamsReturnsDocumentedPagination(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	for _, teamID := range []string{"charlie", "alpha", "bravo"} {
		server.SetTeam("organization", teamID, cm.TeamData{Name: teamID})
	}

	response, err := server.Handler().GetTeams(context.Background(), cm.GetTeamsParams{
		OrganizationID: "organization",
		Skip:           cm.NewOptInt64(1),
		Limit:          cm.NewOptInt64(1),
	})
	require.NoError(t, err)

	page, ok := response.(*cm.TeamCollection)
	require.True(t, ok)
	assert.Equal(t, 3, page.Total.Or(-1))
	assert.Equal(t, 1, page.Skip.Or(-1))
	assert.Equal(t, 1, page.Limit.Or(-1))
	require.Len(t, page.Items, 1)
	assert.Equal(t, "bravo", page.Items[0].Sys.ID)

	beyondEndResponse, err := server.Handler().GetTeams(context.Background(), cm.GetTeamsParams{
		OrganizationID: "organization",
		Skip:           cm.NewOptInt64(999999),
		Limit:          cm.NewOptInt64(1),
	})
	require.NoError(t, err)

	beyondEnd, ok := beyondEndResponse.(*cm.TeamCollection)
	require.True(t, ok)
	assert.Equal(t, 3, beyondEnd.Total.Or(-1))
	assert.Equal(t, 999999, beyondEnd.Skip.Or(-1))
	assert.Equal(t, 1, beyondEnd.Limit.Or(-1))
	assert.Empty(t, beyondEnd.Items)
}

func TestGetTeamsEnforcesUserManagementLimitMaximum(t *testing.T) {
	t.Parallel()

	handler := cmt.NewHandler()
	maximumResponse, err := handler.GetTeams(context.Background(), cm.GetTeamsParams{
		OrganizationID: "organization",
		Limit:          cm.NewOptInt64(100),
	})
	require.NoError(t, err)

	_, ok := maximumResponse.(*cm.TeamCollection)
	require.True(t, ok)

	response, err := handler.GetTeams(context.Background(), cm.GetTeamsParams{
		OrganizationID: "organization",
		Limit:          cm.NewOptInt64(101),
	})
	require.NoError(t, err)

	statusCode, ok := response.(*cm.ErrorStatusCode)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, statusCode.StatusCode)
}
