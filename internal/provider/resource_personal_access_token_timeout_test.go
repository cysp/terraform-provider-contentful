package provider_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

//nolint:gosec // This is a Terraform resource address, not a credential.
const testAccPersonalAccessTokenResourceAddress = "contentful_personal_access_token.test"

func TestAccPersonalAccessTokenResourceMockTimeoutUpdate(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(100))
	require.NoError(t, err)

	counter := &personalAccessTokenMutationCounter{handler: server}
	idSame := statecheck.CompareValue(compare.ValuesSame())
	tokenSame := statecheck.CompareValue(compare.ValuesSame())
	expiresAtSame := statecheck.CompareValue(compare.ValuesSame())
	revokedAtSame := statecheck.CompareValue(compare.ValuesSame())
	preservedStateChecks := func(timeoutCheck knownvalue.Check) []statecheck.StateCheck {
		return []statecheck.StateCheck{
			idSame.AddStateValue(testAccPersonalAccessTokenResourceAddress, tfjsonpath.New("id")),
			tokenSame.AddStateValue(testAccPersonalAccessTokenResourceAddress, tfjsonpath.New("token")),
			expiresAtSame.AddStateValue(testAccPersonalAccessTokenResourceAddress, tfjsonpath.New("expires_at")),
			revokedAtSame.AddStateValue(testAccPersonalAccessTokenResourceAddress, tfjsonpath.New("revoked_at")),
			statecheck.ExpectKnownValue(testAccPersonalAccessTokenResourceAddress, tfjsonpath.New("name"), knownvalue.StringExact("terraform-provider-contentful-acctest-timeouts")),
			statecheck.ExpectKnownValue(testAccPersonalAccessTokenResourceAddress, tfjsonpath.New("scopes"), knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("content_management_read")})),
			statecheck.ExpectKnownValue(testAccPersonalAccessTokenResourceAddress, tfjsonpath.New("expires_in"), knownvalue.Int64Exact(300)),
			statecheck.ExpectKnownValue(testAccPersonalAccessTokenResourceAddress, tfjsonpath.New("timeouts"), timeoutCheck),
			statecheck.ExpectIdentityValueMatchesState(testAccPersonalAccessTokenResourceAddress, tfjsonpath.New("id")),
		}
	}

	steps := make([]resource.TestStep, 0, 4)
	steps = append(steps, resource.TestStep{
		Config:            testAccPersonalAccessTokenConfig(""),
		ConfigStateChecks: preservedStateChecks(knownvalue.Null()),
	})

	for _, timeoutChange := range []struct {
		config string
		check  knownvalue.Check
	}{
		{
			config: `timeouts = { create = "1m", read = "2m", delete = "3m" }`,
			check: knownvalue.ObjectExact(map[string]knownvalue.Check{
				"create": knownvalue.StringExact("1m"),
				"read":   knownvalue.StringExact("2m"),
				"delete": knownvalue.StringExact("3m"),
			}),
		},
		{
			config: `timeouts = { create = "4m", read = "5m", delete = "6m" }`,
			check: knownvalue.ObjectExact(map[string]knownvalue.Check{
				"create": knownvalue.StringExact("4m"),
				"read":   knownvalue.StringExact("5m"),
				"delete": knownvalue.StringExact("6m"),
			}),
		},
		{config: "", check: knownvalue.Null()},
	} {
		steps = append(steps, resource.TestStep{
			Config: testAccPersonalAccessTokenConfig(timeoutChange.config),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(testAccPersonalAccessTokenResourceAddress, plancheck.ResourceActionUpdate),
				},
				PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
			},
			ConfigStateChecks: preservedStateChecks(timeoutChange.check),
		})
	}

	ContentfulProviderMockedResourceTest(t, counter, resource.TestCase{
		Steps: steps,
	})

	require.Equal(t, int64(1), counter.creates.Load(), "timeout updates must not create another token")
	require.Equal(t, int64(1), counter.revocations.Load(), "only test cleanup may revoke the token")
}

func TestAccPersonalAccessTokenResourceMockImportedTimeoutUpdate(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(100))
	require.NoError(t, err)

	const importedResourceName = "terraform-provider-contentful-acctest-imported-timeouts"

	response, err := server.Handler().CreatePersonalAccessToken(context.Background(), &cm.PersonalAccessTokenRequestData{
		Name:   importedResourceName,
		Scopes: []string{"content_management_read"},
	})
	require.NoError(t, err)

	created, ok := response.(*cm.PersonalAccessTokenStatusCode)
	require.True(t, ok, "unexpected personal access token creation response: %T", response)

	counter := &personalAccessTokenMutationCounter{handler: server}

	ContentfulProviderMockedResourceTest(t, counter, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccImportedPersonalAccessTokenConfig(created.Response.Sys.ID, importedResourceName, ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				Config: testAccImportedPersonalAccessTokenConfig(created.Response.Sys.ID, importedResourceName, `timeouts = { create = "7m", read = "8m", delete = "9m" }`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccPersonalAccessTokenResourceAddress, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(testAccPersonalAccessTokenResourceAddress, tfjsonpath.New("id"), knownvalue.StringExact(created.Response.Sys.ID)),
					statecheck.ExpectKnownValue(testAccPersonalAccessTokenResourceAddress, tfjsonpath.New("name"), knownvalue.StringExact(importedResourceName)),
					statecheck.ExpectKnownValue(testAccPersonalAccessTokenResourceAddress, tfjsonpath.New("scopes"), knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("content_management_read")})),
					statecheck.ExpectKnownValue(testAccPersonalAccessTokenResourceAddress, tfjsonpath.New("token"), knownvalue.Null()),
					statecheck.ExpectKnownValue(testAccPersonalAccessTokenResourceAddress, tfjsonpath.New("expires_in"), knownvalue.Null()),
					statecheck.ExpectKnownValue(testAccPersonalAccessTokenResourceAddress, tfjsonpath.New("expires_at"), knownvalue.Null()),
					statecheck.ExpectKnownValue(testAccPersonalAccessTokenResourceAddress, tfjsonpath.New("revoked_at"), knownvalue.Null()),
					statecheck.ExpectKnownValue(testAccPersonalAccessTokenResourceAddress, tfjsonpath.New("timeouts"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"create": knownvalue.StringExact("7m"),
						"read":   knownvalue.StringExact("8m"),
						"delete": knownvalue.StringExact("9m"),
					})),
					statecheck.ExpectIdentityValueMatchesState(testAccPersonalAccessTokenResourceAddress, tfjsonpath.New("id")),
				},
			},
		},
	})

	require.Equal(t, int64(0), counter.creates.Load(), "imported timeout update must not create a token")
	require.Equal(t, int64(1), counter.revocations.Load(), "only test cleanup may revoke the imported token")
}

func testAccPersonalAccessTokenConfig(timeouts string) string {
	return fmt.Sprintf(`
resource "contentful_personal_access_token" "test" {
  name       = "terraform-provider-contentful-acctest-timeouts"
  scopes     = ["content_management_read"]
  expires_in = 5 * 60

  %s
}
`, timeouts)
}

func testAccImportedPersonalAccessTokenConfig(tokenID, resourceName, timeouts string) string {
	return fmt.Sprintf(`
import {
  id = %[1]q
  to = contentful_personal_access_token.test
}

resource "contentful_personal_access_token" "test" {
  name   = %[2]q
  scopes = ["content_management_read"]

  %[3]s
}
`, tokenID, resourceName, timeouts)
}

type personalAccessTokenMutationCounter struct {
	handler     http.Handler
	creates     atomic.Int64
	revocations atomic.Int64
}

func (c *personalAccessTokenMutationCounter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.URL.Path == "/users/me/access_tokens" && request.Method == http.MethodPost {
		c.creates.Add(1)
	}

	if strings.HasPrefix(request.URL.Path, "/users/me/access_tokens/") && strings.HasSuffix(request.URL.Path, "/revoked") && request.Method == http.MethodPut {
		c.revocations.Add(1)
	}

	c.handler.ServeHTTP(responseWriter, request)
}
