package provider_test

import (
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccWebhookResourceImportedTimeoutOnlyUpdate(t *testing.T) {
	t.Parallel()

	const (
		address = "contentful_webhook.test"
		timeout = `timeouts = { update = "3m" }`
	)

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	fields := cm.WebhookDefinitionData{
		Name: "Before import", URL: "https://example.com/webhook", Topics: []string{"Entry.publish"},
		HttpBasicUsername: cm.NewOptNilString("user"), HttpBasicPassword: cm.NewOptNilString("remote-password"),
		Active: cm.NewOptBool(true),
		Headers: cm.WebhookDefinitionHeaders{
			{Key: "X-Ordinary", Secret: cm.NewOptBool(false)},
			{Key: "X-Managed", Value: cm.NewOptString("same"), Secret: cm.NewOptBool(false)},
		},
	}
	server.SetWebhookDefinition("space", "imported-webhook", fields)
	// Advance the version before import to exercise the stored PUT precondition.
	fields.Name = "Imported webhook"
	seed, err := server.Handler().UpdateWebhookDefinition(t.Context(), &fields, cm.UpdateWebhookDefinitionParams{
		SpaceID: "space", WebhookDefinitionID: "imported-webhook", XContentfulVersion: cm.NewOptInt(0),
	})
	require.NoError(t, err)
	require.IsType(t, &cm.WebhookDefinitionStatusCode{}, seed)

	recorder := &taxonomyRequestBodyRecorder{next: server}
	// The configuration omits both credentials. The effective plan retains the imported username.
	localConfig := webhookBasicPasswordConfig("Imported webhook", "null", timeout+"\nlifecycle { ignore_changes = [http_basic_username] }")
	testAccMockedResource(t, recorder, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				Config: localConfig, ResourceName: address,
				ImportState: true, ImportStateId: "space/imported-webhook", ImportStatePersist: true,
			},
			{
				Config: localConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate),
					plancheck.ExpectKnownValue(address, tfjsonpath.New("version"), knownvalue.Int64Exact(1)),
				}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(address, tfjsonpath.New("version"), knownvalue.Int64Exact(1)),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("timeouts").AtMapKey("update"), knownvalue.StringExact("3m")),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("http_basic_username"), knownvalue.StringExact("user")),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("http_basic_password"), knownvalue.Null()),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("headers").AtMapKey("X-Ordinary"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"secret": knownvalue.Bool(false), "value": knownvalue.Null(),
					})),
					statecheck.ExpectIdentityValueMatchesState(address, tfjsonpath.New("space_id")),
					statecheck.ExpectIdentityValueMatchesState(address, tfjsonpath.New("webhook_id")),
				},
				Check: func(*terraform.State) error {
					require.Empty(t, recorder.matchingRequests(http.MethodPut, ""))

					stored, ok := server.StoredWebhookDefinition("space", "imported-webhook")
					require.True(t, ok)
					require.Equal(t, 1, stored.Sys.Version)
					require.Equal(t, cm.NewOptNilString("remote-password"), stored.HttpBasicPassword)

					return nil
				},
			},
			{
				Config: webhookBasicPasswordConfig("Imported webhook", "null", `
  timeouts = { update = "4m" }
  headers  = { X-Managed = { value = "same" } }
  lifecycle { ignore_changes = [http_basic_username, headers["X-Ordinary"]] }`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(address, tfjsonpath.New("timeouts").AtMapKey("update"), knownvalue.StringExact("4m")),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("headers").AtMapKey("X-Ordinary").AtMapKey("value"), knownvalue.Null()),
				},
			},
			{
				Config: webhookBasicPasswordConfig("Imported webhook", `"rotated-password"`, timeout+"\nheaders = {}"),
				Check: func(*terraform.State) error {
					puts := recorder.matchingRequests(http.MethodPut, "")
					require.Len(t, puts, 1)
					require.Equal(t, "/spaces/space/webhook_definitions/imported-webhook", puts[0].path)
					require.Equal(t, "1", puts[0].version)
					require.JSONEq(t, testJSON(map[string]any{
						"name": "Imported webhook", "url": "https://example.com/webhook", "topics": []string{"Entry.publish"},
						"httpBasicUsername": "user", "httpBasicPassword": "rotated-password",
						"filters": nil, "headers": []any{}, "transformation": nil, "active": true,
					}), string(puts[0].body))

					stored, ok := server.StoredWebhookDefinition("space", "imported-webhook")
					require.True(t, ok)
					require.Equal(t, 2, stored.Sys.Version)
					require.Equal(t, cm.NewOptNilString("rotated-password"), stored.HttpBasicPassword)

					return nil
				},
			},
		},
	})
}
