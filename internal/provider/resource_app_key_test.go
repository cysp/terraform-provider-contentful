package provider_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

const (
	testAccAppKeyOrganizationID  = "2zuSjSO4A0e6GKBrhJRe2m"
	testAccAppKeyAppDefinitionID = "2fxGxOcam8Fo5m1wC11fhn"
)

func TestAccAppKeyResourceJWK(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name: "Test App",
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id":   config.StringVariable("organization-id"),
					"app_definition_id": config.StringVariable("app-definition-id"),
					"key_kid":           config.StringVariable("key-id"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_app_key.test", "id", "organization-id/app-definition-id/key-id"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "key_kid", "key-id"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.alg", "RS256"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.kty", "RSA"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.use", "sig"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.kid", "key-id"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.x5c.#", "1"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.x5c.0", "certificate"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.x5t", "key-id"),
					resource.TestCheckNoResourceAttr("contentful_app_key.test", "private_key"),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id":   config.StringVariable("organization-id"),
					"app_definition_id": config.StringVariable("app-definition-id"),
					"key_kid":           config.StringVariable("replacement-key-id"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_app_key.test", plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func TestAccAppKeyResourceGenerate(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name: "Test App",
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id":   config.StringVariable("organization-id"),
					"app_definition_id": config.StringVariable("app-definition-id"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_app_key.test", "id", "organization-id/app-definition-id/generated-key"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "key_kid", "generated-key"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.kid", "generated-key"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "private_key", "generated-private-key"),
				),
			},
			{
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_app_key.test", "private_key", "generated-private-key"),
				),
			},
		},
	})
}

//nolint:paralleltest
func TestAccAppKeyResourceLiveGenerate(t *testing.T) {
	parallelWhenMocked(t)

	cleanupLiveAppKeyFixture(t)

	server, _ := cmt.NewContentfulManagementServer()
	server.SetAppDefinition(testAccAppKeyOrganizationID, testAccAppKeyAppDefinitionID, cm.AppDefinitionData{
		Name: "Test App",
	})

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccAppKeyResourceLiveGeneratedConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_app_key.test", "organization_id", testAccAppKeyOrganizationID),
					resource.TestCheckResourceAttr("contentful_app_key.test", "app_definition_id", testAccAppKeyAppDefinitionID),
					resource.TestCheckResourceAttrSet("contentful_app_key.test", "key_kid"),
					resource.TestCheckResourceAttrSet("contentful_app_key.test", "jwk.kid"),
					resource.TestCheckResourceAttrSet("contentful_app_key.test", "private_key"),
				),
			},
			{
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("contentful_app_key.test", "private_key"),
				),
			},
		},
	})
}

//nolint:paralleltest
func TestAccAppKeyResourceLiveJWK(t *testing.T) {
	parallelWhenMocked(t)

	cleanupLiveAppKeyFixture(t)

	server, _ := cmt.NewContentfulManagementServer()
	server.SetAppDefinition(testAccAppKeyOrganizationID, testAccAppKeyAppDefinitionID, cm.AppDefinitionData{
		Name: "Test App",
	})

	keyKID, x5t, x5c := testAccAppKeyPublicKeyJWK(t)

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccAppKeyResourceLiveJWKConfig(keyKID, x5t, x5c),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_app_key.test", "organization_id", testAccAppKeyOrganizationID),
					resource.TestCheckResourceAttr("contentful_app_key.test", "app_definition_id", testAccAppKeyAppDefinitionID),
					resource.TestCheckResourceAttr("contentful_app_key.test", "key_kid", keyKID),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.alg", "RS256"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.kty", "RSA"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.use", "sig"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.kid", keyKID),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.x5c.#", "1"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.x5c.0", x5c),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.x5t", x5t),
					resource.TestCheckNoResourceAttr("contentful_app_key.test", "private_key"),
				),
			},
		},
	})
}

func TestAccAppKeyResourceReplaceGeneratedWithSameKIDJWK(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name: "Test App",
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccAppKeyResourceGeneratedConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_app_key.test", "id", "organization-id/app-definition-id/generated-key"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "private_key", "generated-private-key"),
				),
			},
			{
				Config: testAccAppKeyResourceGeneratedKeyIDJWKConfig(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_app_key.test", plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_app_key.test", "id", "organization-id/app-definition-id/generated-key"),
					resource.TestCheckNoResourceAttr("contentful_app_key.test", "private_key"),
				),
			},
		},
	})
}

func TestAccAppKeyResourceGenerateMissingPrivateKey(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name: "Test App",
	})
	server.OmitGeneratedAppKeyPrivateKey()

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      testAccAppKeyResourceGeneratedConfig(),
				ExpectError: regexp.MustCompile(`(?s)did not include a decodable generated private key`),
			},
		},
	})
}

func TestAccAppKeyResourceInvalidJWK(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `
resource "contentful_app_key" "test" {
  organization_id   = "organization-id"
  app_definition_id = "app-definition-id"

  jwk = {
    kid = ""
    x5c = ["certificate"]
    x5t = "key-id"
  }
}
`,
				ExpectError: regexp.MustCompile(`jwk\.kid.*length`),
			},
			{
				Config: `
resource "contentful_app_key" "test" {
  organization_id   = "organization-id"
  app_definition_id = "app-definition-id"

  jwk = {
    kid = "key-id"
    x5c = [""]
    x5t = "key-id"
  }
}
`,
				ExpectError: regexp.MustCompile(`jwk\.x5c.*length`),
			},
			{
				Config: `
resource "contentful_app_key" "test" {
  organization_id   = "organization-id"
  app_definition_id = "app-definition-id"

  jwk = {
    kid = "key-id"
    x5c = ["certificate"]
    x5t = ""
  }
}
`,
				ExpectError: regexp.MustCompile(`jwk\.x5t.*length`),
			},
		},
	})
}

func TestAccAppKeyResourceRemoveJWK(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name: "Test App",
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccAppKeyResourceJWKConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_app_key.test", "id", "organization-id/app-definition-id/key-id"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "key_kid", "key-id"),
					resource.TestCheckNoResourceAttr("contentful_app_key.test", "private_key"),
				),
			},
			{
				Config: testAccAppKeyResourceGeneratedConfig(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_app_key.test", plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_app_key.test", "id", "organization-id/app-definition-id/generated-key"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "key_kid", "generated-key"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.kid", "generated-key"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "private_key", "generated-private-key"),
				),
			},
		},
	})
}

func TestAccAppKeyResourceNullJWK(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name: "Test App",
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_app_key.test", "id", "organization-id/app-definition-id/generated-key"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "key_kid", "generated-key"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.kid", "generated-key"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "private_key", "generated-private-key"),
				),
			},
		},
	})
}

func TestAccAppKeyResourceImport(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"organization_id":   config.StringVariable("organization-id"),
		"app_definition_id": config.StringVariable("app-definition-id"),
		"key_kid":           config.StringVariable("key-id"),
	}

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name: "Test App",
	})

	server.SetAppKey("organization-id", "app-definition-id", cm.AppKeyRequestData{
		Jwk: cm.NewOptAppKeyJWK(testAppKeyJWK("key-id")),
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_app_key.test", "id", "organization-id/app-definition-id/key-id"),
				),
			},
			{
				ConfigDirectory:   config.TestNameDirectory(),
				ConfigVariables:   configVariables,
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      "contentful_app_key.test",
			},
		},
	})
}

func testAccAppKeyResourceJWKConfig() string {
	return `
resource "contentful_app_key" "test" {
  organization_id   = "organization-id"
  app_definition_id = "app-definition-id"

  jwk = {
    kid = "key-id"
    x5c = ["certificate"]
    x5t = "key-id"
  }
}
`
}

func testAccAppKeyResourceGeneratedConfig() string {
	return `
resource "contentful_app_key" "test" {
  organization_id   = "organization-id"
  app_definition_id = "app-definition-id"
}
`
}

func testAccAppKeyResourceGeneratedKeyIDJWKConfig() string {
	return `
resource "contentful_app_key" "test" {
  organization_id   = "organization-id"
  app_definition_id = "app-definition-id"

  jwk = {
    kid = "generated-key"
    x5c = ["certificate"]
    x5t = "generated-key"
  }
}
`
}

func testAccAppKeyResourceLiveGeneratedConfig() string {
	return fmt.Sprintf(`
resource "contentful_app_key" "test" {
  organization_id   = %[1]q
  app_definition_id = %[2]q
}
`, testAccAppKeyOrganizationID, testAccAppKeyAppDefinitionID)
}

func testAccAppKeyResourceLiveJWKConfig(keyKID, x5t, x5c string) string {
	return fmt.Sprintf(`
resource "contentful_app_key" "test" {
  organization_id   = %[1]q
  app_definition_id = %[2]q

  jwk = {
    kid = %[3]q
    x5c = [%[5]q]
    x5t = %[4]q
  }
}
`, testAccAppKeyOrganizationID, testAccAppKeyAppDefinitionID, keyKID, x5t, x5c)
}

func testAccAppKeyPublicKeyJWK(t *testing.T) (string, string, string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Fatalf("failed to generate app key test RSA key: %s", err)
	}

	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal app key test public key: %s", err)
	}

	thumbprint := sha256.Sum256(publicKeyDER)
	keyID := base64.RawURLEncoding.EncodeToString(thumbprint[:])

	return keyID, keyID, base64.StdEncoding.EncodeToString(publicKeyDER)
}

func cleanupLiveAppKeyFixture(t *testing.T) {
	t.Helper()

	if os.Getenv("TF_ACC_MOCKED") != "" {
		return
	}

	accessToken := os.Getenv("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN")
	if accessToken == "" {
		return
	}

	cleanup := func() {
		t.Helper()

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		httpClient := retryablehttp.NewClient()
		httpClient.HTTPClient = http.DefaultClient
		httpClient.Logger = nil
		httpClient.RetryMax = 5

		client, err := cm.NewClient(
			cm.DefaultServerURL,
			cm.NewAccessTokenSecuritySource(accessToken),
			cm.WithClient(cm.NewTransportClient(httpClient.StandardClient(), cm.DefaultUserAgent)),
		)
		if err != nil {
			t.Fatalf("failed to create app key cleanup client: %s", err)
		}

		params := cm.GetAppKeysParams{
			OrganizationID:  testAccAppKeyOrganizationID,
			AppDefinitionID: testAccAppKeyAppDefinitionID,
		}

		response, err := client.GetAppKeys(ctx, params)
		if err != nil {
			t.Fatalf("failed to list existing app keys for fixture cleanup: %s", err)
		}

		keys, ok := response.(*cm.AppKeyCollection)
		if !ok {
			if response, ok := response.(cm.StatusCodeResponse); ok && response.GetStatusCode() == http.StatusNotFound {
				t.Fatalf("app key fixture app definition %q was not found in organization %q", testAccAppKeyAppDefinitionID, testAccAppKeyOrganizationID)
			}

			t.Fatalf("unexpected app key fixture cleanup response: %T", response)
		}

		for _, key := range keys.Items {
			deleteResponse, err := client.DeleteAppKey(ctx, cm.DeleteAppKeyParams{
				OrganizationID:  testAccAppKeyOrganizationID,
				AppDefinitionID: testAccAppKeyAppDefinitionID,
				KeyKid:          key.Sys.ID,
			})
			if err != nil {
				t.Fatalf("failed to delete existing app key %q for fixture cleanup: %s", key.Sys.ID, err)
			}

			if response, ok := deleteResponse.(cm.StatusCodeResponse); ok {
				switch response.GetStatusCode() {
				case http.StatusNoContent, http.StatusNotFound:
				default:
					t.Fatalf("unexpected status deleting existing app key %q for fixture cleanup: %d", key.Sys.ID, response.GetStatusCode())
				}
			}
		}
	}

	cleanup()
	t.Cleanup(cleanup)
}

func testAppKeyJWK(keyID string) cm.AppKeyJWK {
	return cm.AppKeyJWK{
		Alg: cm.AppKeyJWKAlgRS256,
		Kty: cm.AppKeyJWKKtyRSA,
		Use: cm.AppKeyJWKUseSig,
		X5c: []string{"certificate"},
		Kid: keyID,
		X5t: keyID,
	}
}
