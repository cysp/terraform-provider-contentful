package provider_test

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"
	"testing"

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
			)

			name, version := "Original", 7
			missingReadVersion := false
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mutex.Lock()
				defer mutex.Unlock()

				w.Header().Set("Content-Type", "application/json")

				versionJSON := fmt.Sprintf(`,"version":%d`, version)

				switch r.Method {
				case http.MethodPost:
					w.WriteHeader(http.StatusCreated)
				case http.MethodGet:
				case http.MethodDelete:
					w.WriteHeader(http.StatusNoContent)

					return
				case http.MethodPut:
					assert.Equal(t, "/spaces/space/environments/environment/locales/locale", r.URL.Path)

					body, err := io.ReadAll(r.Body)
					assert.NoError(t, err)
					assert.JSONEq(t, `{"name":"Planned","code":"en-AU","fallbackCode":null,"contentDeliveryApi":true,"contentManagementApi":true,"optional":false}`, string(body))

					versions = append(versions, r.Header.Get("X-Contentful-Version"))

					if len(versions) == 1 {
						name, version = "Returned", 9

						versionJSON = `,"version":9`
						if missingVersion {
							versionJSON = ""
						}
					} else {
						name, version = "Planned", 10
						versionJSON = `,"version":10`
					}
				default:
					http.Error(w, "unexpected request", http.StatusBadRequest)

					return
				}

				responseName := name
				if r.Method == http.MethodGet && missingReadVersion {
					responseName, versionJSON = "Remote drift", ""
				}

				_, _ = fmt.Fprintf(w, `{"name":%q,"code":"en-AU","fallbackCode":null,"contentDeliveryApi":true,"contentManagementApi":true,"optional":false,"default":false,"sys":{"type":"Locale","id":"locale","space":{"sys":{"type":"Link","linkType":"Space","id":"space"}},"environment":{"sys":{"type":"Link","linkType":"Environment","id":"environment"}}%s}}`, responseName, versionJSON)
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
			// Check persisted state before refresh can repair it.
			priorStateChecks := func(name string) resource.ConfigPlanChecks {
				return resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					testAccPriorState{check: statecheck.ExpectKnownValue(address, tfjsonpath.New("name"), knownvalue.StringExact(name))},
					testAccPriorState{check: statecheck.ExpectKnownValue(address, tfjsonpath.New("id"), knownvalue.StringExact("space/environment/locale"))},
				}}
			}

			steps := []resource.TestStep{{Config: configuration("Original")}}
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
						Config:           configuration("Planned"),
						ConfigPlanChecks: priorStateChecks("Returned"),
						ExpectError:      regexp.MustCompile("Private version is unavailable"),
					},
					resource.TestStep{
						PreConfig: func() {
							mutex.Lock()
							assert.Equal(t, []string{"7"}, versions, "missing private version must stop before a second mutation")
							mutex.Unlock()
						},
						RefreshState:       true,
						ExpectNonEmptyPlan: true, // The returned name still differs from configuration.
					},
				)
			}

			steps = append(steps, resource.TestStep{
				Config:           configuration("Planned"),
				ConfigPlanChecks: priorStateChecks("Returned"),
			})
			testAccMockedResource(t, handler, resource.TestCase{
				AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
				Steps:                steps,
			})
			mutex.Lock()
			assert.Equal(t, []string{"7", "9"}, versions)
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

	name, optional, version := "Original", false, 1
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()

		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)

			return
		}

		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		}

		if r.Method == http.MethodPut {
			data, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			requests = append(requests, string(data))

			assert.Equal(t, "2", r.Header.Get("X-Contentful-Version"))

			optional, version = true, 3
		}

		_, _ = fmt.Fprintf(w, `{"name":%q,"code":"en-AU","fallbackCode":null,"contentDeliveryApi":true,"contentManagementApi":true,"optional":%t,"default":false,"sys":{"type":"Locale","id":"locale","space":{"sys":{"type":"Link","linkType":"Space","id":"space"}},"environment":{"sys":{"type":"Link","linkType":"Environment","id":"environment"}},"version":%d}}`, name, optional, version)
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
		{Config: configuration(false)},
		{
			PreConfig: func() {
				mutex.Lock()
				name, version = "External", 2
				mutex.Unlock()
			},
			Config: configuration(true),
		},
	}})
	mutex.Lock()
	defer mutex.Unlock()

	require.Len(t, requests, 1)
	assert.JSONEq(t, `{"name":"External","code":"en-AU","fallbackCode":null,"contentDeliveryApi":true,"contentManagementApi":true,"optional":true}`, requests[0])
}
