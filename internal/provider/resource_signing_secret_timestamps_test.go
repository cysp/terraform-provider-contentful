package provider_test

import (
	"io"
	"maps"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

type signingSecretTimestampCase struct {
	resourceType string
	endpoint     string
	id           string
	sys          map[string]any
	priorState   map[string]any
	config       func(value, settings string) string
}

func signingSecretTimestampCases() []signingSecretTimestampCase {
	return []signingSecretTimestampCase{
		{
			resourceType: "contentful_app_signing_secret",
			endpoint:     "/organizations/organization/app_definitions/app/signing_secret",
			id:           "organization/app",
			sys:          map[string]any{"type": "AppSigningSecret", "organization": cm.NewOrganizationLink("organization"), "appDefinition": cm.NewAppDefinitionLink("app")},
			priorState:   map[string]any{"id": "organization/app", "organization_id": "organization", "app_definition_id": "app"},
			config:       appSigningSecretTimeoutConfig,
		},
		{
			resourceType: "contentful_webhook_signing_secret",
			endpoint:     "/spaces/space/webhook_settings/signing_secret",
			id:           "space",
			sys:          map[string]any{"type": "WebhookSigningSecret", "space": cm.NewSpaceLink("space")},
			priorState:   map[string]any{"id": "space", "space_id": "space"},
			config:       func(value, settings string) string { return webhookSigningSecretConfig("space", value, settings) },
		},
	}
}

// Construct raw response JSON to exercise the fractional seconds on the wire.
func (c signingSecretTimestampCase) response(createdAt, updatedAt string) string {
	sys := maps.Clone(c.sys)
	if createdAt != "" {
		sys["createdAt"] = createdAt
	}

	if updatedAt != "" {
		sys["updatedAt"] = updatedAt
	}

	return testJSON(map[string]any{"sys": sys, "redactedValue": "opaque"})
}

func TestAccSigningSecretResourceTimestamps(t *testing.T) {
	t.Parallel()

	for _, test := range signingSecretTimestampCases() {
		t.Run(test.resourceType, func(t *testing.T) {
			t.Parallel()

			const (
				firstCreated = "2026-09-01T00:00:00.123456789Z"
				firstUpdated = "2026-09-01T01:00:00.456Z"
				nextCreated  = "2026-09-01T00:00:00.789Z"
				nextUpdated  = "2026-09-01T01:00:00.987654321Z"
			)

			initial, rotated := strings.Repeat("a", 64), strings.Repeat("b", 64)

			var (
				puts        atomic.Int32
				observation atomic.Value
			)

			handler := http.NewServeMux()
			handler.HandleFunc("PUT "+test.endpoint, func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)

				var (
					status            int
					wantValue, result string
				)

				switch puts.Add(1) {
				case 1:
					status, wantValue = http.StatusCreated, initial
					result = test.response("2026-09-01T10:00:00.123456789+10:00", firstUpdated)
				case 2:
					status, wantValue = http.StatusOK, rotated
					result = test.response(nextCreated, nextUpdated)
				default:
					http.Error(w, "unexpected signing secret PUT", http.StatusBadRequest)

					return
				}

				if err != nil || string(body) != testJSON(map[string]any{"value": wantValue}) || r.URL.RawQuery != "" {
					http.Error(w, "unexpected signing secret request", http.StatusBadRequest)

					return
				}

				observation.Store(result)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				_, _ = io.WriteString(w, result)
			})
			handler.HandleFunc("GET "+test.endpoint, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				body, ok := observation.Load().(string)
				if !ok {
					http.Error(w, "unexpected GET before creation", http.StatusBadRequest)

					return
				}

				_, _ = io.WriteString(w, body)
			})
			handler.HandleFunc("DELETE "+test.endpoint, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })

			address := test.resourceType + ".test"
			created, updated := tfjsonpath.New("created_at"), tfjsonpath.New("updated_at")
			mutationPlan := func(createdAt, updatedAt string) resource.ConfigPlanChecks {
				return resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue(address, created),
						plancheck.ExpectUnknownValue(address, updated),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(address, created, knownvalue.StringExact(createdAt)),
						plancheck.ExpectKnownValue(address, updated, knownvalue.StringExact(updatedAt)),
					},
				}
			}
			timeoutPlan := func(createdAt, updatedAt knownvalue.Check) resource.ConfigPlanChecks {
				return resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate),
					plancheck.ExpectKnownValue(address, created, createdAt),
					plancheck.ExpectKnownValue(address, updated, updatedAt),
				}}
			}
			checks := func(createdAt, updatedAt knownvalue.Check) []statecheck.StateCheck {
				return []statecheck.StateCheck{
					statecheck.ExpectKnownValue(address, created, createdAt),
					statecheck.ExpectKnownValue(address, updated, updatedAt),
				}
			}
			testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
				{Config: test.config(initial, ""), ConfigPlanChecks: mutationPlan(firstCreated, firstUpdated)},
				{Config: test.config(rotated, ""), ConfigPlanChecks: mutationPlan(nextCreated, nextUpdated)},
				{ResourceName: address, ImportState: true, ImportStateId: test.id, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"value"}},
				{Config: test.config(rotated, `timeouts = { read = "1m" }`), ConfigPlanChecks: timeoutPlan(knownvalue.StringExact(nextCreated), knownvalue.StringExact(nextUpdated))},
				{
					PreConfig: func() { observation.Store(test.response("", firstUpdated)) },
					Config: test.config(initial, `timeouts = { read = "2m" }
lifecycle { ignore_changes = [value] }`),
					ConfigPlanChecks:  timeoutPlan(knownvalue.Null(), knownvalue.StringExact(firstUpdated)),
					ConfigStateChecks: append(checks(knownvalue.Null(), knownvalue.StringExact(firstUpdated)), statecheck.ExpectKnownValue(address, tfjsonpath.New("value"), knownvalue.StringExact(rotated))),
				},
				{
					PreConfig:         func() { observation.Store(test.response("", "")) },
					Config:            test.config(rotated, ""),
					ConfigPlanChecks:  timeoutPlan(knownvalue.Null(), knownvalue.Null()),
					ConfigStateChecks: checks(knownvalue.Null(), knownvalue.Null()),
				},
			}})
			require.EqualValues(t, 2, puts.Load())
		})
	}
}

func TestSigningSecretRefreshFromStateWithoutTimestamps(t *testing.T) {
	t.Parallel()

	for _, test := range signingSecretTimestampCases() {
		t.Run(test.resourceType, func(t *testing.T) {
			t.Parallel()

			const created = "2026-09-01T00:00:00.123456Z"

			var calls atomic.Int32

			server := configuredProtocolServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)

				if r.Method != http.MethodGet || r.URL.RequestURI() != test.endpoint {
					http.Error(w, "unexpected signing secret request during refresh", http.StatusBadRequest)

					return
				}

				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, test.response(created, ""))
			}))
			state := test.priorState
			state["value"], state["timeouts"] = strings.Repeat("a", 64), nil
			upgrade, err := server.UpgradeResourceState(t.Context(), &tfprotov6.UpgradeResourceStateRequest{
				TypeName: test.resourceType,
				Version:  0,
				RawState: &tfprotov6.RawState{JSON: []byte(testJSON(state))},
			})
			require.NoError(t, err)
			require.Empty(t, upgrade.Diagnostics)
			read, err := server.ReadResource(t.Context(), &tfprotov6.ReadResourceRequest{TypeName: test.resourceType, CurrentState: upgrade.UpgradedState})
			require.NoError(t, err)
			require.Empty(t, read.Diagnostics)
			require.EqualValues(t, 1, calls.Load())

			schemas, err := server.GetProviderSchema(t.Context(), &tfprotov6.GetProviderSchemaRequest{})
			require.NoError(t, err)

			wireType := schemas.ResourceSchemas[test.resourceType].ValueType()
			state["created_at"], state["updated_at"] = created, nil
			want := tfprotov6.DynamicValue{JSON: []byte(testJSON(state))}
			expected, err := want.Unmarshal(wireType)
			require.NoError(t, err)
			actual, err := read.NewState.Unmarshal(wireType)
			require.NoError(t, err)
			require.True(t, actual.Equal(expected), "refreshed state must retain the ordinary secret and observe the optional timestamp")
		})
	}
}
