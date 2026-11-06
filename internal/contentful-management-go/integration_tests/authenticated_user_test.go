package integration_tests_test

import (
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAuthenticatedUser(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		accessToken string
		userID      string
		status      int
		errorID     string
	}{
		"found":        {accessToken: cmt.ValidAccessToken, userID: "123"},
		"not found":    {accessToken: cmt.ValidAccessToken, status: http.StatusNotFound, errorID: "NotFound"},
		"unauthorized": {accessToken: "CFPAT-00000", status: http.StatusUnauthorized, errorID: "AccessTokenInvalid"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, testServer := testContentfulManagementHTTPTestServer(t)
			client := testContentfulManagementClient(t, testServer.URL, test.accessToken)

			if test.userID != "" {
				server.SetMe(cm.NewUser(test.userID))
			}

			response, err := client.GetAuthenticatedUser(t.Context())
			require.NoError(t, err)
			require.NotNil(t, response)

			if test.errorID != "" {
				if test.status == http.StatusUnauthorized {
					// The generated client uses a dedicated variant for HTTP 401.
					require.IsType(t, &cm.ApplicationJSONError{}, response)
				} else {
					failure, ok := response.(cm.StatusCodeResponse)
					require.True(t, ok, "expected a StatusCodeResponse, got %T", response)
					assert.Equal(t, test.status, failure.GetStatusCode())
				}

				failure, ok := response.(cm.ErrorResponse)
				require.True(t, ok, "expected an ErrorResponse, got %T", response)
				errorResponse, ok := failure.GetError()
				require.True(t, ok)
				assert.Equal(t, test.errorID, errorResponse.Sys.ID)

				return
			}

			user, ok := response.(*cm.User)
			require.True(t, ok, "expected a User, got %T", response)
			assert.Equal(t, test.userID, user.Sys.ID)
		})
	}
}
