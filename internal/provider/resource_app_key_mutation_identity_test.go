package provider_test

import (
	"bytes"
	"io"
	"net/http"
	"regexp"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccAppKeyResourceIdentityContradictionRetainsSuppliedKey(t *testing.T) {
	t.Parallel()

	server := newTestCMAServer(t)
	server.SetAppDefinition("org", "app", cm.AppDefinitionData{Name: "App"})

	requested := testAccAppKeyJWKFromDER(bytes.Repeat([]byte{1}, 600))
	unrelated := testAccAppKeyJWKFromDER(bytes.Repeat([]byte{2}, 600))
	server.SetAppKey("org", "app", cm.AppKeyRequestData{Jwk: cm.AppKeyJWK{
		Alg: "RS256", Kty: "RSA", Use: "sig", Kid: unrelated.kid, X5t: unrelated.x5t, X5c: []string{unrelated.x5c},
	}})

	adapter := &mutationJSONResponseAdapter{delegate: server}
	adapter.mutateNext(http.MethodPost, func(response map[string]any) {
		sys, ok := response["sys"].(map[string]any)
		if assert.True(t, ok) {
			sys["id"] = unrelated.kid
		}
	})

	var (
		requestMutex sync.Mutex
		requests     []string
		postBody     []byte
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestMutex.Lock()

		requests = append(requests, r.Method+" "+r.URL.Path)
		requestMutex.Unlock()

		if r.Method == http.MethodPost {
			body, readErr := io.ReadAll(r.Body)
			assert.NoError(t, readErr)

			r.Body = io.NopCloser(bytes.NewReader(body))

			requestMutex.Lock()
			postBody = body
			requestMutex.Unlock()
		}

		adapter.ServeHTTP(w, r)
	})
	config := testAccAppKeyConfig("org", "app", requested, "")

	testAccMockedResource(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("Contentful returned a different app key identity"),
			},
			{
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPreRefresh: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(testAccAppKeyResourceAddress, plancheck.ResourceActionDestroyBeforeCreate),
					testAccPriorState{check: statecheck.ExpectKnownValue(testAccAppKeyResourceAddress, tfjsonpath.New("id"), knownvalue.StringExact("org/app/"+requested.kid))},
					testAccPriorState{check: statecheck.ExpectKnownValue(testAccAppKeyResourceAddress, tfjsonpath.New("key_kid"), knownvalue.StringExact(requested.kid))},
					testAccPriorState{check: statecheck.ExpectIdentity(testAccAppKeyResourceAddress, map[string]knownvalue.Check{
						"organization_id": knownvalue.StringExact("org"), "app_definition_id": knownvalue.StringExact("app"), "key_kid": knownvalue.StringExact(requested.kid),
					})},
				}},
			},
			{Config: config, Destroy: true},
		},
		CheckDestroy: func(*terraform.State) error {
			requestMutex.Lock()
			assert.Equal(t, []string{
				"POST /organizations/org/app_definitions/app/keys",
				"DELETE /organizations/org/app_definitions/app/keys/" + requested.kid,
			}, requests, "cleanup must address the supplied key without a refresh")
			assert.JSONEq(t, testJSON(map[string]any{"jwk": map[string]any{
				"alg": "RS256", "kty": "RSA", "use": "sig", "kid": requested.kid, "x5t": requested.x5t, "x5c": []string{requested.x5c},
			}}), string(postBody))
			requestMutex.Unlock()

			remaining, readErr := server.Handler().GetAppKey(t.Context(), cm.GetAppKeyParams{OrganizationID: "org", AppDefinitionID: "app", KeyKid: unrelated.kid})
			require.NoError(t, readErr)

			key, ok := remaining.(*cm.AppKey)
			require.True(t, ok, "the unrelated key must survive cleanup")
			assert.Equal(t, unrelated.kid, key.Jwk.Kid)
			assert.Equal(t, []string{unrelated.x5c}, key.Jwk.X5c)

			removed, readErr := server.Handler().GetAppKey(t.Context(), cm.GetAppKeyParams{OrganizationID: "org", AppDefinitionID: "app", KeyKid: requested.kid})
			require.NoError(t, readErr)

			missing, ok := removed.(*cm.ErrorStatusCode)
			require.True(t, ok, "cleanup must delete the supplied key")
			assert.Equal(t, http.StatusNotFound, missing.StatusCode)

			return nil
		},
	})
}
