package provider_test

import (
	"errors"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	testingresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

var errUnexpectedRegistryV0062UpgradeEvidence = errors.New("unexpected Registry v0.0.62 upgrade evidence")

func TestAccContentTypeResourceRegistryV0062ImportRemainsObservational(t *testing.T) {
	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	contentTypeID := "registry-v0062-import"
	draft := seedContentTypeDraft(t, server, contentTypeID)
	require.Equal(t, 1, draft.Sys.Version)
	require.False(t, draft.Sys.PublishedVersion.IsSet())

	handler := &contentTypeActivationTestHandler{delegate: server}
	testserver := httptest.NewServer(handler)
	t.Cleanup(testserver.Close)
	t.Setenv("CONTENTFUL_URL", testserver.URL)
	t.Setenv("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN", "CFPAT-12345")
	t.Setenv(testingresource.EnvTfAccProviderNamespace, "cysp")

	variables := config.Variables{
		"space_id":             config.StringVariable("space"),
		"environment_id":       config.StringVariable("environment"),
		"test_content_type_id": config.StringVariable(contentTypeID),
	}
	additionalCLIOptions := &testingresource.AdditionalCLIOptions{}
	workingDirectoryParent := t.TempDir()
	options := ContentfulProviderOptionsWithHTTPTestServer(testserver)

	testingresource.Test(t, testingresource.TestCase{
		IsUnitTest:           true,
		WorkingDir:           workingDirectoryParent,
		AdditionalCLIOptions: additionalCLIOptions,
		Steps: []testingresource.TestStep{
			{
				ExternalProviders: map[string]testingresource.ExternalProvider{
					"contentful": {
						Source:            "registry.terraform.io/cysp/contentful",
						VersionConstraint: "= 0.0.62",
					},
				},
				Config:             registryV0062ImportContentTypeConfig,
				ConfigVariables:    variables,
				ResourceName:       "contentful_content_type.test",
				ImportState:        true,
				ImportStateId:      "space/environment/" + contentTypeID,
				ImportStatePersist: true,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("%w: expected one imported resource state, got %d", errUnexpectedRegistryV0062UpgradeEvidence, len(states))
					}

					if states[0].Attributes["name"] != "Test" {
						return fmt.Errorf("%w: imported name %q", errUnexpectedRegistryV0062UpgradeEvidence, states[0].Attributes["name"])
					}

					if _, exists := states[0].Attributes["published_version"]; exists {
						return fmt.Errorf("%w: v0.0.62 import wrote candidate-only published_version", errUnexpectedRegistryV0062UpgradeEvidence)
					}

					lockFiles, globErr := filepath.Glob(filepath.Join(workingDirectoryParent, "work*", ".terraform.lock.hcl"))
					if globErr != nil {
						return fmt.Errorf("locate Terraform dependency lock file: %w", globErr)
					}

					if len(lockFiles) != 1 {
						return fmt.Errorf("%w: expected one Terraform dependency lock file, got %d", errUnexpectedRegistryV0062UpgradeEvidence, len(lockFiles))
					}

					lockContents, readErr := os.ReadFile(lockFiles[0])
					if readErr != nil {
						return fmt.Errorf("read Terraform dependency lock file: %w", readErr)
					}

					if !regexp.MustCompile(`provider "registry\.terraform\.io/cysp/contentful"[\s\S]*version\s*=\s*"0\.0\.62"`).Match(lockContents) {
						return fmt.Errorf("%w: Terraform dependency lock file does not select v0.0.62", errUnexpectedRegistryV0062UpgradeEvidence)
					}

					additionalCLIOptions.Plan.NoRefresh = true

					handler.resetRequestHistory()

					return nil
				},
			},
			{
				ProtoV6ProviderFactories: makeTestAccProtoV6ProviderFactories(options...),
				Config:                   registryV0062ContentTypeConfig,
				ConfigVariables:          variables,
				ConfigPlanChecks: testingresource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("published_version"), knownvalue.Null()),
					plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
				}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("published_version"), knownvalue.Null()),
				},
				Check: contentTypeActivationRequestCheck(handler, 0, 0),
			},
			{
				PreConfig: func() {
					additionalCLIOptions.Plan.NoRefresh = false

					handler.resetRequestHistory()
				},
				ProtoV6ProviderFactories: makeTestAccProtoV6ProviderFactories(options...),
				Config:                   registryV0062ContentTypeConfig,
				ConfigVariables:          variables,
				ConfigPlanChecks: testingresource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("published_version"), knownvalue.Null()),
					plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
				}},
				Check: contentTypeActivationRequestCheck(handler, 0, 0),
			},
		},
	})
}

const registryV0062ImportContentTypeConfig = `
terraform {
  required_providers {
    contentful = {
      source  = "cysp/contentful"
      version = "= 0.0.62"
    }
  }
}
` + registryCandidateContentTypeConfig

const registryV0062ContentTypeConfig = `
terraform {
  required_providers {
    contentful = {
      source = "cysp/contentful"
    }
  }
}
` + registryCandidateContentTypeConfig

const registryCandidateContentTypeConfig = `
variable "space_id" {
  type = string
}

variable "environment_id" {
  type = string
}

variable "test_content_type_id" {
  type = string
}

resource "contentful_content_type" "test" {
  space_id        = var.space_id
  environment_id  = var.environment_id
  content_type_id = var.test_content_type_id

  name          = "Test"
  description   = "Test content type (${var.test_content_type_id})"
  display_field = "name"

  fields = [
    {
      id        = "name"
      name      = "Name"
      type      = "Symbol"
      required  = true
      localized = false
    },
    {
      id   = "flags"
      name = "Flags"
      type = "Array"
      items = {
        type = "Symbol"
      }
      required  = false
      localized = false
    },
  ]
}
`

func seedContentTypeDraft(t *testing.T, server *cmt.Server, contentTypeID string) cm.ContentType {
	t.Helper()

	request := cm.ContentTypeRequestData{
		Name:         "Test",
		Description:  cm.NewOptNilString("Test content type (" + contentTypeID + ")"),
		DisplayField: "name",
		Fields: []cm.ContentTypeRequestDataFieldsItem{
			{
				ID: "name", Name: "Name", Type: "Symbol",
				Required: cm.NewOptBool(true), Localized: cm.NewOptBool(false),
				Disabled: cm.NewOptBool(false), Omitted: cm.NewOptBool(false),
			},
			{
				ID: "flags", Name: "Flags", Type: "Array",
				Items:    cm.NewOptContentTypeRequestDataFieldsItemItems(cm.ContentTypeRequestDataFieldsItemItems{Type: cm.NewOptString("Symbol")}),
				Required: cm.NewOptBool(false), Localized: cm.NewOptBool(false),
				Disabled: cm.NewOptBool(false), Omitted: cm.NewOptBool(false),
			},
		},
	}
	putResponse, err := server.Handler().PutContentType(t.Context(), &request, cm.PutContentTypeParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: contentTypeID,
	})
	require.NoError(t, err)

	put, ok := putResponse.(*cm.ContentTypeStatusCode)
	require.True(t, ok)

	return put.Response
}
