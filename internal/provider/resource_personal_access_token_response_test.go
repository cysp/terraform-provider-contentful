package provider_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccPersonalAccessTokenResourceMissingCreateToken(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	var (
		mutex       sync.Mutex
		createdID   string
		createCount int
		fixtureErr  error
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/users/me/access_tokens" {
			server.ServeHTTP(w, r)

			return
		}

		recorder := httptest.NewRecorder()
		server.ServeHTTP(recorder, r)

		mutex.Lock()
		defer mutex.Unlock()

		createCount++

		var response map[string]json.RawMessage

		fixtureErr = json.Unmarshal(recorder.Body.Bytes(), &response)
		if fixtureErr != nil {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		var sys struct {
			ID string `json:"id"`
		}

		fixtureErr = json.Unmarshal(response["sys"], &sys)
		if fixtureErr != nil {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		createdID = sys.ID

		delete(response, "token")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(recorder.Code)
		fixtureErr = json.NewEncoder(w).Encode(response)
	})

	// Omission is an adversarial response, not a claim about normal CMA Create.
	// The identity still represents a created token that Terraform must revoke.
	testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{{
		Config: testAccPersonalAccessTokenConfig(""),
		ConfigStateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue(testAccPersonalAccessTokenResourceAddress, tfjsonpath.New("token"), knownvalue.Null()),
			statecheck.ExpectIdentityValueMatchesState(testAccPersonalAccessTokenResourceAddress, tfjsonpath.New("id")),
		},
	}}})

	mutex.Lock()
	defer mutex.Unlock()

	require.NoError(t, fixtureErr)
	require.Equal(t, 1, createCount)
	require.NotEmpty(t, createdID)

	response, err := server.Handler().GetPersonalAccessToken(t.Context(), cm.GetPersonalAccessTokenParams{AccessTokenID: createdID})
	require.NoError(t, err)

	token, ok := response.(*cm.PersonalAccessToken)
	require.True(t, ok)
	require.True(t, token.RevokedAt.IsSet())
	require.False(t, token.RevokedAt.IsNull())
}
