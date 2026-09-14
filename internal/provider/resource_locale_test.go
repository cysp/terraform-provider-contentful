package provider_test

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

var (
	errLocaleStillExists = errors.New("locale still exists after destroy")
	errLocaleReusedID    = errors.New("recreated locale retained the deleted system ID")
)

func TestAccLocaleResourceCreateUpdateDelete(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	server.RegisterSpaceEnvironment("space-id", "master")

	var localeID string

	identityCheck := statecheck.CompareValue(compare.ValuesSame())

	configVariables := config.Variables{
		"space_id":       config.StringVariable("space-id"),
		"environment_id": config.StringVariable("master"),
		"name":           config.StringVariable("German"),
		"code":           config.StringVariable("de-DE"),
		"fallback_code":  config.StringVariable("en-US"),
	}

	configVariables2 := maps.Clone(configVariables)
	configVariables2["name"] = config.StringVariable("German (Germany)")
	configVariables2["optional"] = config.BoolVariable(true)
	configVariables3 := maps.Clone(configVariables2)
	delete(configVariables3, "fallback_code")
	configVariables3["content_delivery_api"] = config.BoolVariable(false)
	configVariables3["content_management_api"] = config.BoolVariable(false)
	configVariables3["code"] = config.StringVariable("de-AT")

	testAccMockedResource(t, server, resource.TestCase{
		CheckDestroy: func(_ *terraform.State) error {
			response, err := server.Handler().GetLocale(context.WithoutCancel(t.Context()), cm.GetLocaleParams{SpaceID: "space-id", EnvironmentID: "master", LocaleID: localeID})
			if err != nil {
				return fmt.Errorf("read destroyed locale: %w", err)
			}

			status, ok := response.(cm.StatusCodeResponse)
			if !ok || status.GetStatusCode() != http.StatusNotFound {
				return errLocaleStillExists
			}

			return nil
		},
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ConfigStateChecks: []statecheck.StateCheck{
					identityCheck.AddStateValue("contentful_locale.test", tfjsonpath.New("locale_id")),
					statecheck.ExpectKnownValue("contentful_locale.test", tfjsonpath.New("default"), knownvalue.Bool(false)),
				},
				Check: resource.TestCheckResourceAttrWith("contentful_locale.test", "locale_id", func(value string) error {
					localeID = value

					return nil
				}),
			},
			{
				ConfigDirectory:   config.TestNameDirectory(),
				ConfigVariables:   configVariables2,
				ConfigStateChecks: []statecheck.StateCheck{identityCheck.AddStateValue("contentful_locale.test", tfjsonpath.New("locale_id"))},
				ConfigPlanChecks:  resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("contentful_locale.test", plancheck.ResourceActionUpdate)}},
			},
			{
				ConfigDirectory:   config.TestNameDirectory(),
				ConfigVariables:   configVariables2,
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      "contentful_locale.test",
			},
			{
				ConfigDirectory: config.TestNameDirectory(), ConfigVariables: configVariables2,
				ImportState: true, ImportStateKind: resource.ImportBlockWithResourceIdentity,
				ResourceName: "contentful_locale.test",
			},
			{
				ConfigDirectory: config.TestNameDirectory(), ConfigVariables: configVariables3,
				ConfigStateChecks: []statecheck.StateCheck{
					identityCheck.AddStateValue("contentful_locale.test", tfjsonpath.New("locale_id")),
					statecheck.ExpectKnownValue("contentful_locale.test", tfjsonpath.New("fallback_code"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccLocaleResourceExternalDeletion(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	const configuration = `resource "contentful_locale" "test" {
  space_id = "space"
  environment_id = "environment"
  name = "German"
  code = "de-DE"
 }`

	var localeID string

	testAccMockedResource(t, server, resource.TestCase{Steps: []resource.TestStep{
		{
			Config: configuration,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("contentful_locale.test", tfjsonpath.New("content_delivery_api"), knownvalue.Bool(true)),
				statecheck.ExpectKnownValue("contentful_locale.test", tfjsonpath.New("content_management_api"), knownvalue.Bool(true)),
				statecheck.ExpectKnownValue("contentful_locale.test", tfjsonpath.New("optional"), knownvalue.Bool(false)),
			},
			Check: resource.TestCheckResourceAttrWith("contentful_locale.test", "locale_id", func(value string) error {
				localeID = value

				return nil
			}),
		},
		{
			Config: configuration, PreConfig: func() {
				response, err := server.Handler().DeleteLocale(t.Context(), cm.DeleteLocaleParams{SpaceID: "space", EnvironmentID: "environment", LocaleID: localeID})
				require.NoError(t, err)
				require.IsType(t, &cm.NoContent{}, response)
			}, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("contentful_locale.test", plancheck.ResourceActionCreate)}},
			Check: resource.TestCheckResourceAttrWith("contentful_locale.test", "locale_id", func(value string) error {
				if value == localeID {
					return errLocaleReusedID
				}

				return nil
			}),
		},
	}})
}
