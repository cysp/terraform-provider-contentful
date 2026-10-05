package provider_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
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

const livePreviewVariablesAddress = "contentful_live_preview_variables.test"

func livePreviewVariablesTestClient(t *testing.T, handler http.Handler) *cm.Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client, err := cm.NewClient(server.URL, cm.NewAccessTokenSecuritySource(cmt.ValidAccessToken), cm.WithClient(cm.NewTransportClient(server.Client(), "acceptance-test")))
	require.NoError(t, err)

	return client
}

func livePreviewVariablesStep(value string) resource.TestStep {
	return resource.TestStep{
		ConfigDirectory: config.StaticDirectory("testdata/TestAccLivePreviewVariablesResource"),
		ConfigVariables: config.Variables{"variables": config.StringVariable(value)},
		ConfigStateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue(livePreviewVariablesAddress, tfjsonpath.New("variables"), knownvalue.StringExact(value)),
			statecheck.ExpectKnownValue(livePreviewVariablesAddress, tfjsonpath.New("timeouts").AtMapKey("read"), knownvalue.StringExact("3m")),
		},
	}
}

func TestAccLivePreviewVariablesResource(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	client := livePreviewVariablesTestClient(t, server)

	var (
		initial = testJSON(map[string]any{
			"empty":     "",
			"global":    "https://preview.invalid/{entry.sys.id}",
			"localized": map[string]any{"en-US": ""},
			"null":      nil,
		})
		removed = testJSON(map[string]any{"localized": map[string]any{}, "null": nil})
	)

	drift := livePreviewVariablesStep(removed)
	drift.PreConfig = func() {
		current, getErr := client.GetLivePreviewVariables(t.Context(), cm.GetLivePreviewVariablesParams{SpaceID: "space", EnvironmentID: "environment"})
		require.NoError(t, getErr)

		document, ok := current.(*cm.LivePreviewVariables)
		require.True(t, ok)
		updated, putErr := client.PutLivePreviewVariables(t.Context(), &cm.LivePreviewVariablesData{Variables: []byte(testJSON(map[string]any{"external": "change"}))}, cm.PutLivePreviewVariablesParams{SpaceID: "space", EnvironmentID: "environment", XContentfulVersion: document.Sys.Version})
		require.NoError(t, putErr)
		require.IsType(t, &cm.LivePreviewVariables{}, updated)
	}
	drift.ConfigPlanChecks = resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
		plancheck.ExpectResourceAction(livePreviewVariablesAddress, plancheck.ResourceActionUpdate),
		testAccPriorState{check: statecheck.ExpectKnownValue(livePreviewVariablesAddress, tfjsonpath.New("variables"), knownvalue.StringExact(testJSON(map[string]any{"external": "change"})))},
	}}
	disappeared := livePreviewVariablesStep(removed)
	disappeared.PreConfig = func() {
		deleted, deleteErr := client.DeleteLivePreviewVariables(t.Context(), cm.DeleteLivePreviewVariablesParams{SpaceID: "space", EnvironmentID: "environment"})
		require.NoError(t, deleteErr)
		require.IsType(t, &cm.NoContent{}, deleted)
	}
	disappeared.ConfigPlanChecks = resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(livePreviewVariablesAddress, plancheck.ResourceActionCreate)}}
	destroy := livePreviewVariablesStep(testJSON(map[string]any{}))
	destroy.Destroy = true
	destroy.ConfigStateChecks = nil

	testAccMockedResource(t, server, resource.TestCase{Steps: []resource.TestStep{
		livePreviewVariablesStep(initial),
		// Import has no configured timeouts; lifecycle state checks verify their preservation separately.
		{ConfigDirectory: config.StaticDirectory("testdata/TestAccLivePreviewVariablesResource"), ConfigVariables: config.Variables{"variables": config.StringVariable(initial)}, ResourceName: livePreviewVariablesAddress, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"timeouts"}},
		{ConfigDirectory: config.StaticDirectory("testdata/TestAccLivePreviewVariablesResourceImport"), ConfigVariables: config.Variables{"variables": config.StringVariable(initial)}, ResourceName: livePreviewVariablesAddress, ImportState: true, ImportStateKind: resource.ImportBlockWithResourceIdentity},
		livePreviewVariablesStep(removed),
		drift,
		disappeared,
		livePreviewVariablesStep(testJSON(map[string]any{})),
		destroy,
		livePreviewVariablesStep(testJSON(map[string]any{})),
	}})
	read, err := client.GetLivePreviewVariables(t.Context(), cm.GetLivePreviewVariablesParams{SpaceID: "space", EnvironmentID: "environment"})
	require.NoError(t, err)

	failure, ok := read.(*cm.LivePreviewVariablesErrorStatusCode)
	require.True(t, ok, "destroy must remove the document, not leave an empty one")
	require.Equal(t, http.StatusNotFound, failure.StatusCode)
}

func TestAccLivePreviewVariablesResourceIgnoreChanges(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	var (
		mutex  sync.Mutex
		writes []string
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			body, readErr := io.ReadAll(r.Body)
			assert.NoError(t, readErr)
			mutex.Lock()

			writes = append(writes, string(body))
			mutex.Unlock()

			r.Body = io.NopCloser(strings.NewReader(string(body)))
		}

		server.ServeHTTP(w, r)
	})
	initial := livePreviewVariablesStep(testJSON(map[string]any{"managed": "original"}))
	initial.ConfigDirectory = config.StaticDirectory("testdata/TestAccLivePreviewVariablesResourceIgnoreChanges")
	ignored := initial
	ignored.ConfigVariables = config.Variables{"variables": config.StringVariable(testJSON(map[string]any{"managed": "ignored-config"})), "read_timeout": config.StringVariable("4m")}
	ignored.ConfigStateChecks = []statecheck.StateCheck{
		statecheck.ExpectKnownValue(livePreviewVariablesAddress, tfjsonpath.New("variables"), knownvalue.StringExact(testJSON(map[string]any{"managed": "original"}))),
		statecheck.ExpectKnownValue(livePreviewVariablesAddress, tfjsonpath.New("timeouts").AtMapKey("read"), knownvalue.StringExact("4m")),
	}
	ignored.ConfigPlanChecks = resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(livePreviewVariablesAddress, plancheck.ResourceActionUpdate)}}
	testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{initial, ignored}})
	mutex.Lock()
	defer mutex.Unlock()

	require.Len(t, writes, 2)

	for _, body := range writes {
		require.JSONEq(t, testJSON(map[string]any{"variables": map[string]any{"managed": "original"}}), body)
	}
}

func TestAccLivePreviewVariablesResourceIdentityReplacement(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	server.RegisterSpaceEnvironment("space", "other")
	server.RegisterSpaceEnvironment("other-space", "other")

	changedEnvironment := livePreviewVariablesStep(testJSON(map[string]any{}))
	changedEnvironment.ConfigVariables["environment_id"] = config.StringVariable("other")
	changedEnvironment.ConfigPlanChecks = resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(livePreviewVariablesAddress, plancheck.ResourceActionReplace)}}
	changedSpace := livePreviewVariablesStep(testJSON(map[string]any{}))
	changedSpace.ConfigVariables["environment_id"] = config.StringVariable("other")
	changedSpace.ConfigVariables["space_id"] = config.StringVariable("other-space")
	changedSpace.ConfigPlanChecks = changedEnvironment.ConfigPlanChecks
	testAccMockedResource(t, server, resource.TestCase{Steps: []resource.TestStep{livePreviewVariablesStep(testJSON(map[string]any{})), changedEnvironment, changedSpace}})
}

func TestAccLivePreviewVariablesResourceRefusesExistingDocument(t *testing.T) {
	t.Parallel()

	for name, variables := range map[string]string{"empty": testJSON(map[string]any{}), "nonempty": testJSON(map[string]any{"external": "owned"})} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "environment")
			client := livePreviewVariablesTestClient(t, server)
			seeded, err := client.PutLivePreviewVariables(t.Context(), &cm.LivePreviewVariablesData{Variables: []byte(variables)}, cm.PutLivePreviewVariablesParams{SpaceID: "space", EnvironmentID: "environment"})
			require.NoError(t, err)
			require.IsType(t, &cm.LivePreviewVariables{}, seeded)

			step := livePreviewVariablesStep(testJSON(map[string]any{"replacement": "must-not-be-sent"}))
			step.ConfigStateChecks = nil
			step.ExpectError = regexp.MustCompile("existing live preview variables document cannot be overwritten")
			testAccMockedResource(t, server, resource.TestCase{Steps: []resource.TestStep{step}})
			read, err := client.GetLivePreviewVariables(t.Context(), cm.GetLivePreviewVariablesParams{SpaceID: "space", EnvironmentID: "environment"})
			require.NoError(t, err)

			document, ok := read.(*cm.LivePreviewVariables)
			require.True(t, ok)
			require.Equal(t, 1, document.Sys.Version)
			require.JSONEq(t, variables, string(document.Variables))
		})
	}
}

func TestAccLivePreviewVariablesResourceInvalidConfigurationAndImport(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct{ variables, importID, message string }{
		"array rejected":      {variables: testJSON(map[string]any{"value": []any{}}), message: "Invalid live preview variable"},
		"missing import":      {variables: testJSON(map[string]any{}), importID: "space/environment", message: "Cannot import non-existent remote object"},
		"malformed import":    {variables: testJSON(map[string]any{}), importID: "space", message: "Invalid import ID"},
		"empty identity part": {variables: testJSON(map[string]any{}), importID: "space/", message: "Invalid import ID"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "environment")

			step := livePreviewVariablesStep(test.variables)
			step.ConfigStateChecks = nil

			step.ExpectError = regexp.MustCompile(test.message)
			if test.importID != "" {
				step.ResourceName = livePreviewVariablesAddress
				step.ImportState = true
				step.ImportStateId = test.importID
			}

			testAccMockedResource(t, server, resource.TestCase{Steps: []resource.TestStep{step}})
		})
	}
}

func TestAccLivePreviewVariablesResourceServiceValidation(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct{ variables, diagnostic string }{
		"unknown locale": {variables: testJSON(map[string]any{"localized": map[string]any{"zz-ZZ": "VALUE_DO_NOT_ECHO"}}), diagnostic: `zz-ZZ`},
		"reserved name":  {variables: testJSON(map[string]any{"__proto__": "VALUE_DO_NOT_ECHO"}), diagnostic: `Invalid request payload JSON format`},
		"too long":       {variables: testJSON(map[string]any{"global": strings.Repeat("a", 50001)}), diagnostic: `Maximum Text length is 50000 characters`},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "environment")
			client := livePreviewVariablesTestClient(t, server)

			initial := testJSON(map[string]any{"valid": "original"})

			rejected := livePreviewVariablesStep(test.variables)
			rejected.ExpectError = regexp.MustCompile(test.diagnostic)
			rejected.ConfigStateChecks = nil
			recovered := livePreviewVariablesStep(initial)
			recovered.ConfigPlanChecks = resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(livePreviewVariablesAddress, plancheck.ResourceActionNoop)}}
			recovered.PreConfig = func() {
				response, getErr := client.GetLivePreviewVariables(t.Context(), cm.GetLivePreviewVariablesParams{SpaceID: "space", EnvironmentID: "environment"})
				require.NoError(t, getErr)

				document, ok := response.(*cm.LivePreviewVariables)
				require.True(t, ok)
				require.Equal(t, 1, document.Sys.Version)
				require.JSONEq(t, initial, string(document.Variables))
			}
			testAccMockedResource(t, server, resource.TestCase{Steps: []resource.TestStep{livePreviewVariablesStep(initial), rejected, recovered}})
		})
	}
}
