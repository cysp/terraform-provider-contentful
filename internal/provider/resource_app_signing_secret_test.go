package provider_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAppSigningSecretValue        = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaAb09+/=_-"
	testAppSigningSecretUpdatedValue = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbZy87-/=_+"
	testAppSigningSecretRemoteValue  = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccQr65+=_/-"
)

func TestAccAppSigningSecretResource(t *testing.T) {
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
					"organization_id":      config.StringVariable("organization-id"),
					"app_definition_id":    config.StringVariable("app-definition-id"),
					"signing_secret_value": config.StringVariable(testAppSigningSecretValue),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.TestCheckResourceAttr("contentful_app_signing_secret.test", "value", testAppSigningSecretValue),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id":      config.StringVariable("organization-id"),
					"app_definition_id":    config.StringVariable("app-definition-id"),
					"signing_secret_value": config.StringVariable(testAppSigningSecretUpdatedValue),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.TestCheckResourceAttr("contentful_app_signing_secret.test", "value", testAppSigningSecretUpdatedValue),
			},
			{
				PreConfig: func() {
					server.SetAppSigningSecret("organization-id", "app-definition-id", cm.AppSigningSecretRequestData{
						Value: testAppSigningSecretRemoteValue,
					})
				},
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id":      config.StringVariable("organization-id"),
					"app_definition_id":    config.StringVariable("app-definition-id"),
					"signing_secret_value": config.StringVariable(testAppSigningSecretUpdatedValue),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_app_signing_secret.test", plancheck.ResourceActionNoop),
					},
				},
				Check: resource.TestCheckResourceAttr("contentful_app_signing_secret.test", "value", testAppSigningSecretUpdatedValue),
			},
		},
	})
}

func TestAccAppSigningSecretResourceRejectsInvalidValueBeforeContentfulMutation(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"too short":         strings.Repeat("s", 63),
		"too long":          strings.Repeat("s", 65),
		"invalid character": strings.Repeat("s", 63) + "!",
	}

	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)
			server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{Name: "Test App"})

			var putCount atomic.Int64

			handler := http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				if request.Method == http.MethodPut && request.URL.Path == "/organizations/organization-id/app_definitions/app-definition-id/signing_secret" {
					putCount.Add(1)
				}

				server.ServeHTTP(responseWriter, request)
			})

			resourceConfig := fmt.Sprintf(`
resource "contentful_app_signing_secret" "test" {
  organization_id   = "organization-id"
  app_definition_id = "app-definition-id"
  value             = %q
}
`, value)

			ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{Steps: []resource.TestStep{{
				Config:      resourceConfig,
				ExpectError: regexp.MustCompile(`Invalid app signing secret value`),
			}}})

			require.Zero(t, putCount.Load())
		})
	}
}

func TestAccAppSigningSecretResourceWriteOnly(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()
	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name: "Test App",
	})

	signingSecretConfig := fmt.Sprintf(`
resource "contentful_app_signing_secret" "test" {
  organization_id   = "organization-id"
  app_definition_id = "app-definition-id"
  value_wo          = %q
}
`, testAppSigningSecretValue)

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: signingSecretConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("contentful_app_signing_secret.test", "value"),
					resource.TestCheckNoResourceAttr("contentful_app_signing_secret.test", "value_wo"),
				),
			},
			{
				Config: signingSecretConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_app_signing_secret.test", plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccAppSigningSecretImport(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"organization_id":      config.StringVariable("organization-id"),
		"app_definition_id":    config.StringVariable("app-definition-id"),
		"signing_secret_value": config.StringVariable(testAppSigningSecretValue),
	}

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name: "Test App",
	})

	server.SetAppSigningSecret("organization-id", "app-definition-id", cm.AppSigningSecretRequestData{
		Value: testAppSigningSecretRemoteValue,
	})

	var putCount atomic.Int64

	handler := http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPut && request.URL.Path == "/organizations/organization-id/app_definitions/app-definition-id/signing_secret" {
			putCount.Add(1)
		}

		server.ServeHTTP(responseWriter, request)
	})

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:    config.TestNameDirectory(),
				ConfigVariables:    configVariables,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ConfigDirectory:    config.TestNameDirectory(),
				ConfigVariables:    configVariables,
				ImportState:        true,
				ImportStateId:      "organization-id/app-definition-id",
				ImportStatePersist: true,
				ResourceName:       "contentful_app_signing_secret.test",
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contentful_app_signing_secret.test",
						tfjsonpath.New("value"),
						knownvalue.Null(),
					),
				},
			},
			{
				PreConfig:       func() { putCount.Store(0) },
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_app_signing_secret.test", plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_app_signing_secret.test", "id", "organization-id/app-definition-id"),
					resource.TestCheckResourceAttr("contentful_app_signing_secret.test", "value", testAppSigningSecretValue),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_app_signing_secret.test", plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})

	require.Equal(t, int64(1), putCount.Load())
}

func TestAccAppSigningSecretImportBlockWritesConfiguredValue(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name: "Test App",
	})
	server.SetAppSigningSecret("organization-id", "app-definition-id", cm.AppSigningSecretRequestData{
		Value: testAppSigningSecretRemoteValue,
	})

	var (
		putBodiesMutex sync.Mutex
		putBodies      [][]byte
		putBodyReadErr error
	)

	handler := http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPut && request.URL.Path == "/organizations/organization-id/app_definitions/app-definition-id/signing_secret" {
			body, err := io.ReadAll(request.Body)
			if err == nil {
				request.Body = io.NopCloser(bytes.NewReader(body))
			}

			putBodiesMutex.Lock()
			if err != nil {
				putBodyReadErr = err
			} else {
				putBodies = append(putBodies, body)
			}
			putBodiesMutex.Unlock()
		}

		server.ServeHTTP(responseWriter, request)
	})

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`
import {
  id = "organization-id/app-definition-id"
  to = contentful_app_signing_secret.test
}

resource "contentful_app_signing_secret" "test" {
  organization_id   = "organization-id"
  app_definition_id = "app-definition-id"
  value             = %q
}
`, testAppSigningSecretValue),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_app_signing_secret.test", plancheck.ResourceActionUpdate),
				},
				PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
			},
			Check: resource.TestCheckResourceAttr("contentful_app_signing_secret.test", "value", testAppSigningSecretValue),
		}},
	})

	putBodiesMutex.Lock()

	recordedPutBodies := append([][]byte(nil), putBodies...)
	readErr := putBodyReadErr
	putBodiesMutex.Unlock()

	require.NoError(t, readErr)
	require.Len(t, recordedPutBodies, 1)
	assert.JSONEq(t, fmt.Sprintf(`{"value":%q}`, testAppSigningSecretValue), string(recordedPutBodies[0]))
}
