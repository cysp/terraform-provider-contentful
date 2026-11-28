package contentfulmanagement_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppAccessTokenRequiresJWT(t *testing.T) {
	t.Parallel()

	transport := &transportClientStub{err: io.ErrUnexpectedEOF}
	client, err := cm.NewClient(cm.DefaultServerURL, cm.NewAccessTokenSecuritySource("management-token"), cm.WithClient(transport))
	require.NoError(t, err)

	params := cm.CreateAppAccessTokenParams{SpaceID: "space", EnvironmentID: "environment", AppDefinitionID: "app"}

	_, err = client.CreateAppAccessToken(t.Context(), params)
	require.Error(t, err)
	require.Nil(t, transport.request)

	_, err = client.CreateAppAccessToken(cm.WithAppJWT(t.Context(), ""), params)
	require.Error(t, err)
	require.Nil(t, transport.request)
}

func TestAppAccessTokenRequestCredentials(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /spaces/space/environments/environment/app_installations/{app}/access_tokens", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer jwt-"+r.PathValue("app"), r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"token":"issued","sys":{"expiresAt":"2030-01-01T00:00:00Z"}}`)
	})
	mux.HandleFunc("GET /users/me", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer management-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"sys":{"type":"User","id":"user","version":1},"email":"","firstName":"","lastName":""}`)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	client, err := cm.NewClient(server.URL, cm.NewAccessTokenSecuritySource("management-token"), cm.WithClient(server.Client()))
	require.NoError(t, err)

	for _, app := range []string{"first", "second"} {
		ctx := cm.WithAppJWT(t.Context(), "jwt-"+app)
		t.Run(app, func(t *testing.T) {
			t.Parallel()

			_, err := client.CreateAppAccessToken(ctx, cm.CreateAppAccessTokenParams{
				SpaceID: "space", EnvironmentID: "environment", AppDefinitionID: app,
			})
			require.NoError(t, err)
		})
	}

	t.Run("management", func(t *testing.T) {
		t.Parallel()

		_, err := client.GetAuthenticatedUser(cm.WithAppJWT(t.Context(), "jwt-first"))
		require.NoError(t, err)
	})
}
