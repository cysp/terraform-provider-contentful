package contentfulmanagementtestserver_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmts "github.com/cysp/terraform-provider-contentful/internal/contentful-management-testserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleMeNotFound(t *testing.T) {
	t.Parallel()

	ts := cmts.NewContentfulManagementTestServer()

	server := ts.Server()
	defer server.Close()

	client := server.Client()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+"/users/me", nil)
	require.NoError(t, err)

	res, err := client.Do(req)
	require.NoError(t, err)

	defer res.Body.Close()

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestHandleMeFound(t *testing.T) {
	t.Parallel()

	//nolint:varnamelen
	ts := cmts.NewContentfulManagementTestServer()

	ts.SetMe(&cm.User{
		Sys: cm.UserSys{
			ID: "me",
		},
		Email:     "email",
		FirstName: "first",
		LastName:  "last",
	})

	server := ts.Server()
	defer server.Close()

	client := server.Client()

	req, reqErr := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+"/users/me", nil)
	require.NoError(t, reqErr)

	res, resErr := client.Do(req)
	require.NoError(t, resErr)

	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	responseBody, responseBodyErr := io.ReadAll(res.Body)
	require.NoError(t, responseBodyErr)

	var user cm.User

	jsonUnmarshalErr := json.Unmarshal(responseBody, &user)
	require.NoError(t, jsonUnmarshalErr)

	assert.Equal(t, "me", user.Sys.ID)
	assert.Equal(t, "email", user.Email)
	assert.Equal(t, "first", user.FirstName)
	assert.Equal(t, "last", user.LastName)
}
