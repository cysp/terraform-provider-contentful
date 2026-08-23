package provider_test

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
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
