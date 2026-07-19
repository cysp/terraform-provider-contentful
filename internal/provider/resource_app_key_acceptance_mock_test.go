package provider_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	errUnexpectedAppKeyCreateCount = errors.New("unexpected App Key create request count")
	errUnexpectedAppKeyDeleteCount = errors.New("unexpected App Key delete request count")
)

func TestAccAppKeyResourceMockLifecycle(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()
	setTestAccAppKeyAppDefinitions(server)

	jwk := testAccAppKeyJWK(t)
	replacementJWK := testAccAppKeyJWK(t)

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		CheckDestroy: testAccAppKeyDestroyCheck(server.Handler().GetAppKey, jwk.kid, replacementJWK.kid),
		Steps: []resource.TestStep{
			{
				Config: testAccAppKeyConfig(testAccAppKeyOrganizationID, testAccAppKeyAppDefinitionID, jwk, testAccAppKeyCreateBeforeDestroyHCL),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccAppKeyResourceAddress, "id", testAccAppKeyOrganizationID+"/"+testAccAppKeyAppDefinitionID+"/"+jwk.kid),
					resource.TestCheckResourceAttr(testAccAppKeyResourceAddress, "key_kid", jwk.kid),
					resource.TestCheckResourceAttr(testAccAppKeyResourceAddress, "jwk.alg", "RS256"),
					resource.TestCheckResourceAttr(testAccAppKeyResourceAddress, "jwk.kty", "RSA"),
					resource.TestCheckResourceAttr(testAccAppKeyResourceAddress, "jwk.use", "sig"),
					resource.TestCheckResourceAttr(testAccAppKeyResourceAddress, "jwk.kid", jwk.kid),
					resource.TestCheckResourceAttr(testAccAppKeyResourceAddress, "jwk.x5c.0", jwk.x5c),
					resource.TestCheckResourceAttr(testAccAppKeyResourceAddress, "jwk.x5t", jwk.x5t),
				),
			},
			{
				Config: testAccAppKeyConfig(testAccAppKeyOrganizationID, testAccAppKeyAppDefinitionID, replacementJWK, testAccAppKeyCreateBeforeDestroyHCL),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccAppKeyResourceAddress, plancheck.ResourceActionReplace),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

func TestAccAppKeyResourceMockParentReplacement(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()
	setTestAccAppKeyAppDefinitions(server)

	jwk := testAccAppKeyJWK(t)

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccAppKeyConfig(testAccAppKeyOrganizationID, testAccAppKeyAppDefinitionID, jwk, ""),
			},
			{
				Config: testAccAppKeyConfig(testAccAppKeyOrganizationID, testAccAppKeyOtherAppDefinitionID, jwk, ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccAppKeyResourceAddress, plancheck.ResourceActionReplace),
					},
				},
			},
			{
				Config: testAccAppKeyConfig(testAccAppKeyOtherOrganizationID, testAccAppKeyThirdAppDefinitionID, jwk, ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccAppKeyResourceAddress, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func TestAccAppKeyResourceMockTimeoutUpdate(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()
	setTestAccAppKeyAppDefinitions(server)
	counter := &appKeyMutationCounter{handler: server}
	jwk := testAccAppKeyJWK(t)

	ContentfulProviderMockedResourceTest(t, counter, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccAppKeyConfig(
					testAccAppKeyOrganizationID,
					testAccAppKeyAppDefinitionID,
					jwk,
					`timeouts = { create = "1m", read = "1m", delete = "1m" }`,
				),
			},
			{
				Config: testAccAppKeyConfig(
					testAccAppKeyOrganizationID,
					testAccAppKeyAppDefinitionID,
					jwk,
					`timeouts = { create = "2m", read = "2m", delete = "2m" }`,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccAppKeyResourceAddress, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: func(_ *terraform.State) error {
					if got := counter.creates.Load(); got != 1 {
						return fmt.Errorf("%w: got %d, want 1 after timeout update", errUnexpectedAppKeyCreateCount, got)
					}

					if got := counter.deletes.Load(); got != 0 {
						return fmt.Errorf("%w: got %d, want 0 after timeout update", errUnexpectedAppKeyDeleteCount, got)
					}

					return nil
				},
			},
			{
				Config: testAccAppKeyConfig(testAccAppKeyOrganizationID, testAccAppKeyAppDefinitionID, jwk, ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccAppKeyResourceAddress, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

func TestAccAppKeyResourceMockImport(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()
	setTestAccAppKeyAppDefinitions(server)

	jwk := testAccAppKeyJWK(t)
	resourceConfig := testAccAppKeyConfig(testAccAppKeyOrganizationID, testAccAppKeyAppDefinitionID, jwk, "")

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{Config: resourceConfig},
			{
				Config:            resourceConfig,
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      testAccAppKeyResourceAddress,
			},
			{
				Config:          resourceConfig,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				ResourceName:    testAccAppKeyResourceAddress,
			},
		},
	})
}

func TestAccAppKeyResourceMockExternalDeletion(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()
	setTestAccAppKeyAppDefinitions(server)

	jwk := testAccAppKeyJWK(t)
	resourceConfig := testAccAppKeyConfig(testAccAppKeyOrganizationID, testAccAppKeyAppDefinitionID, jwk, "")

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{Config: resourceConfig},
			{
				Config: resourceConfig,
				PreConfig: func() {
					response, err := server.Handler().DeleteAppKey(context.Background(), cm.DeleteAppKeyParams{
						OrganizationID:  testAccAppKeyOrganizationID,
						AppDefinitionID: testAccAppKeyAppDefinitionID,
						KeyKid:          jwk.kid,
					})
					if err != nil {
						t.Fatalf("delete app key outside Terraform: %v", err)
					}

					if _, ok := response.(*cm.NoContent); !ok {
						t.Fatalf("unexpected external delete response: %T", response)
					}
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccAppKeyResourceAddress, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccAppKeyResourceMockCreateBeforeDestroyRejectsReusedKey(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()
	setTestAccAppKeyAppDefinitions(server)

	jwk := testAccAppKeyJWK(t)

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccAppKeyConfig(testAccAppKeyOrganizationID, testAccAppKeyAppDefinitionID, jwk, testAccAppKeyCreateBeforeDestroyHCL),
			},
			{
				Config:      testAccAppKeyConfig(testAccAppKeyOrganizationID, testAccAppKeyOtherAppDefinitionID, jwk, testAccAppKeyCreateBeforeDestroyHCL),
				ExpectError: regexp.MustCompile(`The key is already in use`),
			},
		},
	})
}

func TestAccAppKeyResourceMockInvalidJWKMaterial(t *testing.T) {
	t.Parallel()

	invalidX5T := testAccAppKeyJWK(t)
	invalidX5T.x5t = "invalid-thumbprint"

	invalidKID := testAccAppKeyJWK(t)
	invalidKID.kid = "invalid-key-id"

	whitespace := testAccAppKeyJWK(t)
	whitespace.x5c = whitespace.x5c[:100] + "\n" + whitespace.x5c[100:]

	for name, test := range map[string]struct {
		jwk     testAccAppKeyJWKData
		message string
	}{
		"x5t":        {jwk: invalidX5T, message: `Invalid app key JWK x5t`},
		"kid":        {jwk: invalidKID, message: `Invalid app key JWK kid`},
		"whitespace": {jwk: whitespace, message: `Invalid app key JWK x5c`},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, _ := cmt.NewContentfulManagementServer()
			setTestAccAppKeyAppDefinitions(server)
			counter := &appKeyMutationCounter{handler: server}

			ContentfulProviderMockedResourceTest(t, counter, resource.TestCase{
				Steps: []resource.TestStep{{
					Config:      testAccAppKeyConfig(testAccAppKeyOrganizationID, testAccAppKeyAppDefinitionID, test.jwk, ""),
					ExpectError: regexp.MustCompile(test.message),
				}},
			})

			if got := counter.creates.Load(); got != 0 {
				t.Fatalf("invalid configuration made %d create requests, want 0", got)
			}
		})
	}
}

func TestAccAppKeyResourceMockAcceptsFingerprintableMaterial(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()
	setTestAccAppKeyAppDefinitions(server)

	jwk := testAccAppKeyJWKFromDER(bytes.Repeat([]byte{0}, 600))

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{{
			Config: testAccAppKeyConfig(testAccAppKeyOrganizationID, testAccAppKeyAppDefinitionID, jwk, ""),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
			},
		}},
	})
}

func TestAccAppKeyResourceMockPreservesNonCanonicalBase64(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()
	setTestAccAppKeyAppDefinitions(server)

	publicKeyBytes := bytes.Repeat([]byte{0}, 550)

	jwk := testAccAppKeyJWKFromDER(publicKeyBytes)
	if !strings.HasSuffix(jwk.x5c, "AA==") {
		t.Fatalf("canonical test input has unexpected suffix: %q", jwk.x5c[len(jwk.x5c)-4:])
	}

	jwk.x5c = jwk.x5c[:len(jwk.x5c)-3] + "P=="

	decoded, err := base64.StdEncoding.DecodeString(jwk.x5c)
	if err != nil {
		t.Fatalf("decode noncanonical test input: %v", err)
	}

	if !bytes.Equal(publicKeyBytes, decoded) {
		t.Fatal("noncanonical test input decodes to different bytes")
	}

	config := testAccAppKeyConfig(testAccAppKeyOrganizationID, testAccAppKeyAppDefinitionID, jwk, "")

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.TestCheckResourceAttr(testAccAppKeyResourceAddress, "jwk.x5c.0", jwk.x5c),
			},
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.TestCheckResourceAttr(testAccAppKeyResourceAddress, "jwk.x5c.0", jwk.x5c),
			},
		},
	})
}

func TestAccAppKeyResourceMockDefersUnknownJWKValidation(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()
	setTestAccAppKeyAppDefinitions(server)

	jwk := testAccAppKeyJWK(t)

	resourceConfig := fmt.Sprintf(`
resource "terraform_data" "key" {
  input = {
    kid = %q
    x5c = %q
    x5t = %q
  }
}

resource "contentful_app_key" "test" {
  organization_id   = %q
  app_definition_id = %q

  jwk = {
    alg = "RS256"
    kty = "RSA"
    use = "sig"
    kid = terraform_data.key.output.kid
    x5c = [terraform_data.key.output.x5c]
    x5t = terraform_data.key.output.x5t
  }
}
`, jwk.kid, jwk.x5c, jwk.x5t, testAccAppKeyOrganizationID, testAccAppKeyAppDefinitionID)

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{{
			Config: resourceConfig,
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("terraform_data.key", plancheck.ResourceActionCreate),
					plancheck.ExpectResourceAction(testAccAppKeyResourceAddress, plancheck.ResourceActionCreate),
				},
				PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
			},
		}},
	})
}

type appKeyMutationCounter struct {
	handler http.Handler
	creates atomic.Int64
	deletes atomic.Int64
}

func (c *appKeyMutationCounter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if strings.HasSuffix(request.URL.Path, "/keys") && request.Method == http.MethodPost {
		c.creates.Add(1)
	}

	if strings.Contains(request.URL.Path, "/keys/") && request.Method == http.MethodDelete {
		c.deletes.Add(1)
	}

	c.handler.ServeHTTP(responseWriter, request)
}
