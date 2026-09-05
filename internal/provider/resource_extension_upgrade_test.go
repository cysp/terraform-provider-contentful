package provider_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	testingresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

var errUnexpectedExtensionUpgrade = errors.New("unexpected Registry v0.0.62 Extension upgrade")

type registryV0062ExtensionSourceCase struct {
	name         string
	src          cm.OptString
	srcdoc       cm.OptString
	legacySrc    string
	legacySrcdoc string
	wantSrc      knownvalue.Check
	wantSrcdoc   knownvalue.Check
}

//nolint:paralleltest // Subtests change process-wide Terraform provider environment variables.
func TestAccExtensionResourceRegistryV0062UpgradeDoesNotMutateSources(t *testing.T) {
	for _, source := range []registryV0062ExtensionSourceCase{
		{
			name:       "src",
			src:        cm.NewOptString("https://example.com/extension.js"),
			legacySrc:  "https://example.com/extension.js",
			wantSrc:    knownvalue.StringExact("https://example.com/extension.js"),
			wantSrcdoc: knownvalue.Null(),
		},
		{
			name:         "srcdoc",
			srcdoc:       cm.NewOptString("<!doctype html><title>Extension</title>"),
			legacySrcdoc: "<!doctype html><title>Extension</title>",
			wantSrc:      knownvalue.Null(),
			wantSrcdoc:   knownvalue.StringExact("<!doctype html><title>Extension</title>"),
		},
		{
			name:       "empty srcdoc",
			srcdoc:     cm.NewOptString(""),
			wantSrc:    knownvalue.Null(),
			wantSrcdoc: knownvalue.StringExact(""),
		},
	} {
		for _, refresh := range []bool{true, false} {
			mode := "refresh enabled"
			if !refresh {
				mode = "refresh disabled"
			}

			t.Run(source.name+"/"+mode, func(t *testing.T) {
				testAccRegistryV0062ExtensionSourceUpgrade(t, source, refresh)
			})
		}
	}
}

func testAccRegistryV0062ExtensionSourceUpgrade(t *testing.T, source registryV0062ExtensionSourceCase, refresh bool) {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	extensionID := "registry-v0062-" + strings.ReplaceAll(source.name, " ", "-")
	name := "Registry v0.0.62 " + source.name
	_, err = server.Handler().PutExtension(t.Context(), &cm.ExtensionData{
		Extension: cm.ExtensionDataExtension{
			Name:       name,
			Src:        source.src,
			Srcdoc:     source.srcdoc,
			FieldTypes: []cm.ExtensionDataExtensionFieldTypesItem{},
			Sidebar:    cm.NewOptBool(false),
		},
	}, cm.PutExtensionParams{SpaceID: "space", EnvironmentID: "environment", ExtensionID: extensionID})
	require.NoError(t, err)

	recorder := &extensionUpgradeMutationRecorder{next: server}
	testserver := httptest.NewServer(recorder)
	t.Cleanup(testserver.Close)
	t.Setenv("CONTENTFUL_URL", testserver.URL)
	t.Setenv("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN", "CFPAT-12345")
	t.Setenv(testingresource.EnvTfAccProviderNamespace, "cysp")

	currentConfig := fmt.Sprintf(`
resource "contentful_extension" "test" {
  space_id       = "space"
  environment_id = "environment"
  extension_id   = %q

  extension = {
    name        = %q
    field_types = []
  }
}
`, extensionID, name)
	previousConfig := `
terraform {
  required_providers {
    contentful = {
      source  = "cysp/contentful"
      version = "= 0.0.62"
    }
  }
}
` + currentConfig

	additionalCLIOptions := &testingresource.AdditionalCLIOptions{}
	options := ContentfulProviderOptionsWithHTTPTestServer(testserver)
	stateChecks := []statecheck.StateCheck{
		statecheck.ExpectKnownValue("contentful_extension.test", tfjsonpath.New("extension").AtMapKey("src"), source.wantSrc),
		statecheck.ExpectKnownValue("contentful_extension.test", tfjsonpath.New("extension").AtMapKey("srcdoc"), source.wantSrcdoc),
	}
	planChecks := []plancheck.PlanCheck{
		plancheck.ExpectResourceAction("contentful_extension.test", plancheck.ResourceActionNoop),
	}
	checkNoPuts := func(*terraform.State) error {
		if puts := recorder.puts.Load(); puts != 0 {
			return fmt.Errorf("%w: got %d PUT requests", errUnexpectedExtensionUpgrade, puts)
		}

		return nil
	}

	testingresource.Test(t, testingresource.TestCase{
		WorkingDir:           t.TempDir(),
		AdditionalCLIOptions: additionalCLIOptions,
		Steps: []testingresource.TestStep{
			{
				ExternalProviders: map[string]testingresource.ExternalProvider{
					"contentful": {
						Source:            "registry.terraform.io/cysp/contentful",
						VersionConstraint: "= 0.0.62",
					},
				},
				Config:             previousConfig,
				ResourceName:       "contentful_extension.test",
				ImportState:        true,
				ImportStateId:      "space/environment/" + extensionID,
				ImportStatePersist: true,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("%w: expected one imported Extension state, got %d", errUnexpectedExtensionUpgrade, len(states))
					}

					if actual := states[0].Attributes["extension.src"]; actual != source.legacySrc {
						return fmt.Errorf("%w: v0.0.62 imported extension.src %q, want %q", errUnexpectedExtensionUpgrade, actual, source.legacySrc)
					}

					if actual := states[0].Attributes["extension.srcdoc"]; actual != source.legacySrcdoc {
						return fmt.Errorf("%w: v0.0.62 imported extension.srcdoc %q, want %q", errUnexpectedExtensionUpgrade, actual, source.legacySrcdoc)
					}

					recorder.puts.Store(0)

					additionalCLIOptions.Plan.NoRefresh = !refresh

					return nil
				},
			},
			{
				ProtoV6ProviderFactories: makeTestAccProtoV6ProviderFactories(options...),
				Config:                   currentConfig,
				ConfigPlanChecks:         testingresource.ConfigPlanChecks{PreApply: planChecks},
				ConfigStateChecks:        stateChecks,
				Check:                    checkNoPuts,
			},
		},
	})
}

type extensionUpgradeMutationRecorder struct {
	next http.Handler
	puts atomic.Int64
}

func (r *extensionUpgradeMutationRecorder) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodPut && strings.Contains(request.URL.Path, "/extensions/") {
		r.puts.Add(1)
	}

	r.next.ServeHTTP(responseWriter, request)
}
