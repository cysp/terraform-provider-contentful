package cmtesting_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentfulManagementServerGetAuthenticatedUserFound(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	hts := httptest.NewServer(server)
	t.Cleanup(hts.Close)

	client, err := cm.NewClient(
		hts.URL,
		cm.NewAccessTokenSecuritySource(cmt.ValidAccessToken),
		cm.WithClient(hts.Client()),
	)
	require.NoError(t, err)

	server.SetMe(cm.NewUser("user123"))

	res, err := client.GetAuthenticatedUser(t.Context())
	assert.NotNil(t, res)
	require.NoError(t, err)

	user, userOk := res.(*cm.User)
	require.True(t, userOk)
	require.NotNil(t, user)

	assert.Equal(t, "user123", user.Sys.ID)
}

func TestContentfulManagementServerGetAuthenticatedUserNotFound(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	hts := httptest.NewServer(server)
	t.Cleanup(hts.Close)

	client, err := cm.NewClient(
		hts.URL,
		cm.NewAccessTokenSecuritySource(cmt.ValidAccessToken),
		cm.WithClient(hts.Client()),
	)
	require.NoError(t, err)

	res, err := client.GetAuthenticatedUser(t.Context())
	assert.NotNil(t, res)
	require.NoError(t, err)

	esc, escOk := res.(*cm.ErrorStatusCode)
	require.True(t, escOk)
	require.NotNil(t, esc)

	assert.Equal(t, http.StatusNotFound, esc.StatusCode)
}
