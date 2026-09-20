package provider_test

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccLocaleResourceUpdateRecovery(t *testing.T) {
	t.Parallel()

	for _, missingVersion := range []bool{false, true} {
		t.Run(fmt.Sprintf("missing_version=%t", missingVersion), func(t *testing.T) {
			t.Parallel()

			var (
				mutex    sync.Mutex
				versions []string
				requests []string
			)

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "environment")

			var localeID string

			missingReadVersion := false
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mutex.Lock()
				defer mutex.Unlock()

				if r.Method == http.MethodPut {
					assert.Equal(t, "/spaces/space/environments/environment/locales/"+localeID, r.URL.Path)
					body, readErr := io.ReadAll(r.Body)
					assert.NoError(t, readErr)

					r.Body = io.NopCloser(bytes.NewReader(body))
					requests = append(requests, string(body))
					versions = append(versions, r.Header.Get("X-Contentful-Version"))
				}

				response := httptest.NewRecorder()
				server.ServeHTTP(response, r)
				maps.Copy(w.Header(), response.Header())

				raw := response.Body.String()
				if r.Method == http.MethodPut && len(versions) == 1 {
					// The server commits Planned. Only the response contradicts it.
					raw = strings.Replace(raw, `"name":"Planned"`, `"name":"Returned"`, 1)
					if missingVersion {
						raw = strings.Replace(raw, `,"version":2`, "", 1)
					}
				}

				if r.Method == http.MethodGet && missingReadVersion {
					raw = strings.Replace(raw, `"name":"Original"`, `"name":"Remote drift"`, 1)
					raw = strings.Replace(raw, `,"version":1`, "", 1)
				}

				w.WriteHeader(response.Code)
				_, _ = io.WriteString(w, raw)
			})

			configuration := func(name string) string {
				return fmt.Sprintf(`resource "contentful_locale" "test" {
  space_id = "space"
  environment_id = "environment"
  name = %q
  code = "en-AU"
}`, name)
			}

			const address = "contentful_locale.test"

			identity := statecheck.CompareValue(compare.ValuesSame())
			// Check persisted state before refresh can repair it.
			priorStateChecks := func(name string) resource.ConfigPlanChecks {
				return resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					testAccPriorState{check: statecheck.ExpectKnownValue(address, tfjsonpath.New("name"), knownvalue.StringExact(name))},
					testAccPriorState{check: identity.AddStateValue(address, tfjsonpath.New("id"))},
				}}
			}

			steps := []resource.TestStep{{Config: configuration("Original"), ConfigStateChecks: []statecheck.StateCheck{
				identity.AddStateValue(address, tfjsonpath.New("id")),
				statecheck.ExpectKnownValue(address, tfjsonpath.New("locale_id"), knownvalue.StringFunc(func(value string) error {
					mutex.Lock()
					localeID = value
					mutex.Unlock()

					return nil
				})),
			}}}
			if missingVersion {
				steps = append(steps, resource.TestStep{
					PreConfig: func() {
						mutex.Lock()
						missingReadVersion = true
						mutex.Unlock()
					},
					RefreshState: true,
					ExpectError:  regexp.MustCompile("Missing locale version"),
				})
			}

			steps = append(steps, resource.TestStep{
				PreConfig: func() {
					mutex.Lock()
					missingReadVersion = false
					mutex.Unlock()
				},
				Config:           configuration("Planned"),
				ConfigPlanChecks: priorStateChecks("Original"),
				ExpectError:      regexp.MustCompile("Contentful returned a different locale name"),
			})
			if missingVersion {
				steps = append(steps,
					resource.TestStep{
						Config:           configuration("Final"),
						ConfigPlanChecks: priorStateChecks("Returned"),
						ExpectError:      regexp.MustCompile("Private version is unavailable"),
					},
					resource.TestStep{
						PreConfig: func() {
							mutex.Lock()
							assert.Equal(t, []string{"1"}, versions, "missing private version must stop before a second mutation")
							mutex.Unlock()
						},
						RefreshState:       true,
						ExpectNonEmptyPlan: true, // Refresh recovers Planned; configuration still requests Final.
					},
				)
			}

			priorName := "Returned"
			if missingVersion {
				priorName = "Planned"
			}

			steps = append(steps, resource.TestStep{
				Config:           configuration("Final"),
				ConfigPlanChecks: priorStateChecks(priorName),
			})
			testAccMockedResource(t, handler, resource.TestCase{
				AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
				Steps:                steps,
			})
			response, getErr := server.Handler().GetLocale(t.Context(), cm.GetLocaleParams{SpaceID: "space", EnvironmentID: "environment", LocaleID: localeID})
			require.NoError(t, getErr)

			status, ok := response.(*cm.ErrorStatusCode)
			require.True(t, ok)
			assert.Equal(t, http.StatusNotFound, status.StatusCode)

			mutex.Lock()
			assert.Equal(t, []string{"1", "2"}, versions)
			require.Len(t, requests, 2)
			assert.JSONEq(t, `{"name":"Planned","code":"en-AU","fallbackCode":null,"contentDeliveryApi":true,"contentManagementApi":true,"optional":false}`, requests[0])
			assert.JSONEq(t, `{"name":"Final","code":"en-AU","fallbackCode":null,"contentDeliveryApi":true,"contentManagementApi":true,"optional":false}`, requests[1])
			mutex.Unlock()
		})
	}
}

func TestAccLocaleResourceUpdatePreservesIgnoredDrift(t *testing.T) {
	t.Parallel()

	var (
		mutex    sync.Mutex
		requests []string
	)

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	var localeID string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			data, readErr := io.ReadAll(r.Body)
			if readErr != nil {
				http.Error(w, readErr.Error(), http.StatusBadRequest)

				return
			}

			r.Body = io.NopCloser(bytes.NewReader(data))

			mutex.Lock()

			requests = append(requests, string(data))
			mutex.Unlock()
			assert.Equal(t, "2", r.Header.Get("X-Contentful-Version"))
		}

		server.ServeHTTP(w, r)
	})

	configuration := func(optional bool) string {
		return fmt.Sprintf(`resource "contentful_locale" "test" {
  space_id = "space"
  environment_id = "environment"
  name = "Original"
  code = "en-AU"
  optional = %t
  lifecycle { ignore_changes = [name] }
}`, optional)
	}
	testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
		{Config: configuration(false), ConfigStateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("contentful_locale.test", tfjsonpath.New("locale_id"), knownvalue.StringFunc(func(value string) error {
				localeID = value

				return nil
			})),
		}},
		{
			PreConfig: func() {
				response, putErr := server.Handler().PutLocale(t.Context(), &cm.LocaleData{
					Name: "External", Code: "en-AU", FallbackCode: cm.NewNilStringNull(), ContentDeliveryApi: true, ContentManagementApi: true,
				}, cm.PutLocaleParams{SpaceID: "space", EnvironmentID: "environment", LocaleID: localeID, XContentfulVersion: 1})
				require.NoError(t, putErr)

				status, ok := response.(*cm.LocaleStatusCode)
				require.True(t, ok)
				assert.Equal(t, http.StatusOK, status.StatusCode)
			},
			Config:            configuration(true),
			ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectKnownValue("contentful_locale.test", tfjsonpath.New("name"), knownvalue.StringExact("External"))},
		},
	}})
	response, getErr := server.Handler().GetLocale(t.Context(), cm.GetLocaleParams{SpaceID: "space", EnvironmentID: "environment", LocaleID: localeID})
	require.NoError(t, getErr)

	status, ok := response.(*cm.ErrorStatusCode)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, status.StatusCode)

	mutex.Lock()
	defer mutex.Unlock()

	require.Len(t, requests, 1)
	assert.JSONEq(t, `{"name":"External","code":"en-AU","fallbackCode":null,"contentDeliveryApi":true,"contentManagementApi":true,"optional":true}`, requests[0])
}
