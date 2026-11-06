package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoveryDataSourcesRawRequestsAndProjection(t *testing.T) {
	t.Parallel()

	for _, family := range discoveryTestFamilies() {
		t.Run(family.name, func(t *testing.T) {
			t.Parallel()
			body := discoveryFixture(t, family.name)

			inputs := map[string]any{}
			maps.Copy(inputs, family.scopes)

			inputs[family.id] = "item-b"
			count := 0
			response := discoveryReadTest(t.Context(), t, family.singular, inputs, roundTripFunc(func(request *http.Request) (*http.Response, error) {
				count++

				assert.Equal(t, http.MethodGet, request.Method)
				assert.Equal(t, "/base"+family.endpoint+"/item-b", request.URL.Path)
				assert.Empty(t, request.URL.RawQuery)
				assert.Empty(t, request.Header.Get("X-Contentful-Organization"))
				assert.Equal(t, "Bearer synthetic-token", request.Header.Get("Authorization"))
				_, hasDeadline := request.Context().Deadline()
				assert.True(t, hasDeadline)

				return discoveryHTTPResponse(request, http.StatusOK, body), nil
			}))
			require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
			assert.Equal(t, 1, count)

			var expectedFields map[string]json.RawMessage
			require.NoError(t, json.Unmarshal([]byte(family.expected), &expectedFields))

			for name, expected := range expectedFields {
				discoveryStateAttribute(t, response, name, string(expected))
			}

			var fields map[string]tftypes.Value
			require.NoError(t, response.State.Raw.As(&fields))
			assert.NotContains(t, fields, "internal_code")

			if family.name == "locale" {
				discoveryStateAttribute(t, response, "fallback_code", "null")
				discoveryStateAttribute(t, response, "code", testJSON("en-GB"))
			}

			if family.name == "environment" {
				discoveryStateAttribute(t, response, "status", testJSON("queued"))
				discoveryStateAttribute(t, response, "aliased_environment_id", testJSON("target"))
			}
		})
	}
}

//nolint:forcetypeassert // Fixture mutations target independently defined object shapes.
func TestDiscoveryDataSourcesPagination(t *testing.T) {
	t.Parallel()

	for _, family := range discoveryTestFamilies() {
		t.Run(family.name, func(t *testing.T) {
			t.Parallel()
			body := discoveryFixture(t, family.name)

			inputs := family.scopes
			if family.organization != "" {
				inputs["organization_id"] = family.organization
			}

			offsets := []string{}
			response := discoveryReadTest(t.Context(), t, family.plural, inputs, roundTripFunc(func(request *http.Request) (*http.Response, error) {
				assert.Equal(t, http.MethodGet, request.Method)
				assert.Equal(t, "/base"+family.endpoint, request.URL.Path)
				assert.Equal(t, "Bearer synthetic-token", request.Header.Get("Authorization"))
				assert.Equal(t, family.organization, request.Header.Get("X-Contentful-Organization"))
				assert.Equal(t, "100", request.URL.Query().Get("limit"))
				assert.Len(t, request.URL.Query(), 2)
				skip := request.URL.Query().Get("skip")

				offsets = append(offsets, skip)
				if skip == "0" {
					return discoveryHTTPResponse(request, 200, discoveryPage(0, 1, 2, body)), nil
				}

				assert.Equal(t, "1", skip)

				second := mutateTestJSON(body, func(document map[string]any) {
					document["sys"].(map[string]any)["id"] = "item-a"
					if _, exists := document["name"]; exists {
						document["name"] = "First"
					}
				})

				return discoveryHTTPResponse(request, 200, discoveryPage(1, 1, 2, second)), nil
			}))
			require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
			assert.Equal(t, []string{"0", "1"}, offsets)

			first := mutateTestJSON(family.expected, func(document map[string]any) {
				document[family.id] = "item-a"
				if _, exists := document["name"]; exists {
					document["name"] = "First"
				}
			})
			if family.preserveResponseOrder {
				discoveryStateAttribute(t, response, family.collection, testJSON([]any{json.RawMessage(family.expected), json.RawMessage(first)}))
			} else {
				discoveryStateAttribute(t, response, family.collection, testJSON([]any{json.RawMessage(first), json.RawMessage(family.expected)}))
			}
		})
	}
}

func TestDiscoveryDataSourcesPaginationMetadataBoundaries(t *testing.T) {
	t.Parallel()

	for _, family := range discoveryTestFamilies() {
		t.Run(family.name, func(t *testing.T) {
			t.Parallel()
			body := discoveryFixture(t, family.name)

			cases := []struct {
				name      string
				pages     []string
				wantCount int
			}{
				{"without metadata", []string{testJSON(map[string]any{"sys": map[string]any{"type": "Array"}, "items": []any{json.RawMessage(body)}}), testJSON(map[string]any{"sys": map[string]any{"type": "Array"}, "items": []any{}})}, 1},
				{"empty before total and incorrect echoed metadata", []string{discoveryPage(99, 0, 10, body), discoveryPage(99, 0, 10)}, 1},
			}
			for _, testcase := range cases {
				t.Run(testcase.name, func(t *testing.T) {
					t.Parallel()

					count := 0
					response := discoveryReadTest(t.Context(), t, family.plural, family.scopes, roundTripFunc(func(request *http.Request) (*http.Response, error) {
						require.Less(t, count, len(testcase.pages))
						assert.Equal(t, strconv.Itoa(count), request.URL.Query().Get("skip"))
						assert.Equal(t, "100", request.URL.Query().Get("limit"))

						page := testcase.pages[count]
						count++

						return discoveryHTTPResponse(request, http.StatusOK, page), nil
					}))
					require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
					assert.Equal(t, len(testcase.pages), count)

					expected := make([]json.RawMessage, testcase.wantCount)
					for i := range expected {
						expected[i] = json.RawMessage(family.expected)
					}

					discoveryStateAttribute(t, response, family.collection, testJSON(expected))
				})
			}
		})
	}
}

func TestDiscoveryDataSourcesPaginationFailuresPublishNothing(t *testing.T) {
	t.Parallel()

	for _, family := range discoveryTestFamilies() {
		t.Run(family.name, func(t *testing.T) {
			t.Parallel()
			body := discoveryFixture(t, family.name)

			cases := []struct {
				name   string
				status int
				body   string
				want   string
			}{
				{"permission", 403, testJSON(map[string]any{"sys": map[string]any{"type": "Error", "id": "AccessDenied"}, "message": "denied"}), "AccessDenied"},
				{"missing parent", 404, testJSON(map[string]any{"sys": map[string]any{"type": "Error", "id": "NotFound"}}), "NotFound"},
				{"malformed", 200, `{"items":`, "decode"},
			}
			for _, testcase := range cases {
				t.Run(testcase.name, func(t *testing.T) {
					t.Parallel()

					count := 0
					response := discoveryReadTest(t.Context(), t, family.plural, family.scopes, roundTripFunc(func(request *http.Request) (*http.Response, error) {
						count++
						if count == 1 {
							return discoveryHTTPResponse(request, 200, discoveryPage(0, 1, 2, body)), nil
						}

						return discoveryHTTPResponse(request, testcase.status, testcase.body), nil
					}))
					require.True(t, response.Diagnostics.HasError())
					assert.Contains(t, fmt.Sprint(response.Diagnostics), testcase.want)
					assert.True(t, response.State.Raw.IsNull())
					assert.Equal(t, 2, count)
				})
			}
		})
	}
}

func TestDiscoveryDataSourcesEmpty(t *testing.T) {
	t.Parallel()

	for _, family := range discoveryTestFamilies() {
		t.Run(family.name, func(t *testing.T) {
			t.Parallel()
			response := discoveryReadTest(t.Context(), t, family.plural, family.scopes, roundTripFunc(func(request *http.Request) (*http.Response, error) {
				assert.Empty(t, request.Header.Get("X-Contentful-Organization"))

				return discoveryHTTPResponse(request, 200, discoveryPage(0, 100, 0)), nil
			}))
			require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
			discoveryStateAttribute(t, response, family.collection, testJSON([]any{}))

			if family.name == "space" {
				discoveryStateAttribute(t, response, "id", testJSON("spaces"))
			}
		})
	}
}

func TestDiscoveryDataSourcesUnresolvedAndInvalidIDsDoNotRead(t *testing.T) {
	t.Parallel()

	for _, family := range discoveryTestFamilies() {
		t.Run(family.name, func(t *testing.T) {
			t.Parallel()

			for _, value := range []any{nil, tftypes.UnknownValue, "", ".", "..", "a/b"} {
				t.Run(fmt.Sprint(value), func(t *testing.T) {
					t.Parallel()

					inputs := map[string]any{}
					maps.Copy(inputs, family.scopes)

					inputs[family.id] = value
					count := 0
					response := discoveryReadTest(t.Context(), t, family.singular, inputs, roundTripFunc(func(request *http.Request) (*http.Response, error) {
						count++

						return discoveryHTTPResponse(request, 500, ""), nil
					}))
					assert.True(t, response.Diagnostics.HasError())
					assert.True(t, response.State.Raw.IsNull())
					assert.Zero(t, count)
				})
			}
		})
	}
}

func TestDiscoveryDataSourcesCancellationAndTimeout(t *testing.T) {
	t.Parallel()

	for _, family := range discoveryTestFamilies() {
		t.Run(family.name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				count := 0
				body := discoveryFixture(t, family.name)

				ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
				defer cancel()

				inputs := maps.Clone(family.scopes)
				inputs["timeouts"] = map[string]tftypes.Value{"read": tftypes.NewValue(tftypes.String, "1s")}
				started := time.Now()
				response := discoveryReadTest(ctx, t, family.plural, inputs, roundTripFunc(func(request *http.Request) (*http.Response, error) {
					count++
					if count == 1 {
						return discoveryHTTPResponse(request, 200, discoveryPage(0, 1, 2, body)), nil
					}

					<-request.Context().Done()

					return nil, request.Context().Err()
				}))
				assert.True(t, response.Diagnostics.HasError())
				assert.True(t, response.State.Raw.IsNull())
				assert.Equal(t, 2, count)
				assert.Equal(t, time.Second, time.Since(started))
			})
			ctx, cancel := context.WithCancel(t.Context())
			cancel()

			count := 0
			response := discoveryReadTest(ctx, t, family.plural, family.scopes, roundTripFunc(func(request *http.Request) (*http.Response, error) {
				count++

				return discoveryHTTPResponse(request, 200, ""), nil
			}))
			assert.True(t, response.Diagnostics.HasError())
			assert.True(t, response.State.Raw.IsNull())
			assert.Zero(t, count)
		})
	}
}

func TestDiscoveryDataSourcesCancellationAfterResponsePublishesNothing(t *testing.T) {
	t.Parallel()

	for _, family := range discoveryTestFamilies() {
		for _, plural := range []bool{false, true} {
			t.Run(family.name+"/"+strconv.FormatBool(plural), func(t *testing.T) {
				t.Parallel()

				factory := family.singular
				inputs := maps.Clone(family.scopes)
				inputs[family.id] = "item-b"
				body := discoveryFixture(t, family.name)

				name := strings.ReplaceAll(family.name, "_", " ")
				if plural {
					factory = family.plural
					inputs = family.scopes
					body = discoveryPage(0, 100, 1, body)
					name = strings.ReplaceAll(family.collection, "_", " ")
				}

				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()

				count := 0
				response := discoveryReadTest(ctx, t, factory, inputs, roundTripFunc(func(request *http.Request) (*http.Response, error) {
					count++

					cancel()

					return discoveryHTTPResponse(request, http.StatusOK, body), nil
				}))
				require.Len(t, response.Diagnostics, 1)
				assert.Equal(t, "Failed to read "+name, response.Diagnostics[0].Summary())
				assert.Equal(t, context.Canceled.Error(), response.Diagnostics[0].Detail())
				assert.True(t, response.Diagnostics.HasError())
				assert.True(t, response.State.Raw.IsNull())
				assert.Equal(t, 1, count)
			})
		}
	}
}

//nolint:forcetypeassert // Fixture mutations target independently defined object shapes.
func TestDiscoveryDataSourcesResponseIdentityAndShape(t *testing.T) {
	t.Parallel()

	for _, family := range discoveryTestFamilies() {
		t.Run(family.name, func(t *testing.T) {
			t.Parallel()
			body := discoveryFixture(t, family.name)

			cases := []struct {
				name   string
				status int
				body   string
			}{
				{"different ID", 200, mutateTestJSON(body, func(document map[string]any) { document["sys"].(map[string]any)["id"] = "different" })},
				{"empty ID", 200, mutateTestJSON(body, func(document map[string]any) { document["sys"].(map[string]any)["id"] = "" })},
				{"not found", 404, testJSON(map[string]any{"sys": map[string]any{"type": "Error", "id": "NotFound"}})},
				{"unauthorized", 401, testJSON(map[string]any{"sys": map[string]any{"type": "Error", "id": "AccessTokenInvalid"}})},
				{"forbidden", 403, testJSON(map[string]any{"sys": map[string]any{"type": "Error", "id": "AccessDenied"}})},
				{"wrong type", 200, testJSON(map[string]any{"sys": map[string]any{"type": "Array"}, "items": []any{}})},
				{"null body", 200, testJSON(nil)},
				{"trailing body", 200, body + ` {}`},
			}
			if family.name != "space" {
				cases = append(cases, struct {
					name   string
					status int
					body   string
				}{"different space", 200, mutateTestJSON(body, func(document map[string]any) {
					document["sys"].(map[string]any)["space"].(map[string]any)["sys"].(map[string]any)["id"] = "other-space"
				})})
			}

			if family.name == "locale" {
				cases = append(cases, struct {
					name   string
					status int
					body   string
				}{"different environment", 200, mutateTestJSON(body, func(document map[string]any) {
					document["sys"].(map[string]any)["environment"].(map[string]any)["sys"].(map[string]any)["id"] = "concrete"
				})})
			}

			for _, testcase := range cases {
				t.Run(testcase.name, func(t *testing.T) {
					t.Parallel()

					inputs := map[string]any{}
					maps.Copy(inputs, family.scopes)

					inputs[family.id] = "item-b"
					response := discoveryReadTest(t.Context(), t, family.singular, inputs, roundTripFunc(func(request *http.Request) (*http.Response, error) {
						return discoveryHTTPResponse(request, testcase.status, testcase.body), nil
					}))
					assert.True(t, response.Diagnostics.HasError(), response)
					assert.True(t, response.State.Raw.IsNull())
				})
			}
		})
	}
}

//nolint:forcetypeassert // Fixture mutations target independently defined object shapes.
func TestDiscoveryDataSourcesEscapedIDs(t *testing.T) {
	t.Parallel()

	for _, family := range discoveryTestFamilies() {
		t.Run(family.name, func(t *testing.T) {
			t.Parallel()
			body := mutateTestJSON(discoveryFixture(t, family.name), func(document map[string]any) { document["sys"].(map[string]any)["id"] = "a%2Fb?c#d" })

			inputs := map[string]any{}
			maps.Copy(inputs, family.scopes)

			inputs[family.id] = "a%2Fb?c#d"
			response := discoveryReadTest(t.Context(), t, family.singular, inputs, roundTripFunc(func(request *http.Request) (*http.Response, error) {
				assert.Equal(t, "/base"+family.endpoint+"/a%252Fb%3Fc%23d", request.URL.EscapedPath())
				assert.Empty(t, request.URL.RawQuery)
				assert.Empty(t, request.URL.Fragment)

				return discoveryHTTPResponse(request, 200, body), nil
			}))
			require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
		})
	}
}

func TestDiscoveryDataSourcesUnknownParentScopesDoNotRead(t *testing.T) {
	t.Parallel()

	for _, family := range discoveryTestFamilies() {
		t.Run(family.name, func(t *testing.T) {
			t.Parallel()

			inputs := map[string]any{}
			for field := range family.scopes {
				inputs[field] = tftypes.UnknownValue
			}

			if family.name == "space" {
				inputs["organization_id"] = tftypes.UnknownValue
			}

			count := 0
			response := discoveryReadTest(t.Context(), t, family.plural, inputs, roundTripFunc(func(request *http.Request) (*http.Response, error) {
				count++

				return discoveryHTTPResponse(request, 500, ""), nil
			}))
			assert.True(t, response.Diagnostics.HasError())
			assert.True(t, response.State.Raw.IsNull())
			assert.Zero(t, count)
		})
	}
}

//nolint:forcetypeassert // Fixture mutations target independently defined object shapes.
func TestDiscoveryDataSourcesLaterScopeContradictionPublishesNothing(t *testing.T) {
	t.Parallel()

	for _, family := range discoveryTestFamilies() {
		t.Run(family.name, func(t *testing.T) {
			t.Parallel()
			body := discoveryFixture(t, family.name)

			inputs := family.scopes
			if family.name == "space" {
				inputs["organization_id"] = "org"
			}

			second := mutateTestJSON(body, func(document map[string]any) {
				sys := document["sys"].(map[string]any)
				sys["id"] = "item-a"

				scope := "space"
				if family.name == "space" {
					scope = "organization"
				}

				sys[scope].(map[string]any)["sys"].(map[string]any)["id"] = "other"
			})
			count := 0
			response := discoveryReadTest(t.Context(), t, family.plural, inputs, roundTripFunc(func(request *http.Request) (*http.Response, error) {
				count++
				if count == 1 {
					return discoveryHTTPResponse(request, 200, discoveryPage(0, 1, 2, body)), nil
				}

				return discoveryHTTPResponse(request, 200, discoveryPage(1, 1, 2, second)), nil
			}))
			assert.True(t, response.Diagnostics.HasError())
			assert.True(t, response.State.Raw.IsNull())
			assert.Equal(t, 2, count)
		})
	}
}

//nolint:forcetypeassert // Fixture mutations target independently defined object shapes.
func TestDiscoveryCollectionsPreserveIrregularItemIDs(t *testing.T) {
	t.Parallel()

	for _, family := range discoveryTestFamilies() {
		t.Run(family.name, func(t *testing.T) {
			t.Parallel()

			for _, returnedID := range []string{"", ".", "..", "a/b"} {
				t.Run(returnedID, func(t *testing.T) {
					t.Parallel()
					body := discoveryFixture(t, family.name)
					irregular := mutateTestJSON(body, func(document map[string]any) { document["sys"].(map[string]any)["id"] = returnedID })
					count := 0
					response := discoveryReadTest(t.Context(), t, family.plural, family.scopes, roundTripFunc(func(request *http.Request) (*http.Response, error) {
						count++
						if count == 1 {
							assert.Equal(t, "0", request.URL.Query().Get("skip"))

							return discoveryHTTPResponse(request, 200, discoveryPage(0, 100, 3, body)), nil
						}

						require.Equal(t, 2, count)
						assert.Equal(t, "1", request.URL.Query().Get("skip"))

						return discoveryHTTPResponse(request, 200, discoveryPage(1, 100, 3, irregular, irregular)), nil
					}))
					require.Empty(t, response.Diagnostics)
					assert.Equal(t, 2, count)

					expected := mutateTestJSON(family.expected, func(document map[string]any) { document[family.id] = returnedID })
					if family.preserveResponseOrder {
						discoveryStateAttribute(t, response, family.collection, testJSON([]any{json.RawMessage(family.expected), json.RawMessage(expected), json.RawMessage(expected)}))
					} else {
						discoveryStateAttribute(t, response, family.collection, testJSON([]any{json.RawMessage(expected), json.RawMessage(expected), json.RawMessage(family.expected)}))
					}

					var attributes map[string]tftypes.Value
					require.NoError(t, response.State.Raw.As(&attributes))

					var items []tftypes.Value
					require.NoError(t, attributes[family.collection].As(&items))

					itemIndex := 0
					if family.preserveResponseOrder {
						itemIndex = 1
					}

					var item map[string]tftypes.Value
					require.NoError(t, items[itemIndex].As(&item))

					var discoveredID string
					require.NoError(t, item[family.id].As(&discoveredID))

					inputs := maps.Clone(family.scopes)
					inputs[family.id] = discoveredID
					requests := 0
					lookup := discoveryReadTest(t.Context(), t, family.singular, inputs, roundTripFunc(func(request *http.Request) (*http.Response, error) {
						requests++

						return discoveryHTTPResponse(request, 500, ""), nil
					}))
					require.True(t, lookup.Diagnostics.HasError())
					assert.Contains(t, fmt.Sprint(lookup.Diagnostics), "Invalid lookup ID")
					assert.Zero(t, requests)
				})
			}
		})
	}
}

//nolint:forcetypeassert // Fixture mutations target independently defined object shapes.
func TestDiscoveryDataSourcesPreserveComputedReferences(t *testing.T) {
	t.Parallel()

	for _, family := range discoveryTestFamilies() {
		if family.name == "locale" || family.name == "role" {
			continue
		}

		t.Run(family.name, func(t *testing.T) {
			t.Parallel()

			for _, returnedID := range []string{"", ".", "..", "a/b"} {
				t.Run(returnedID, func(t *testing.T) {
					t.Parallel()
					body := mutateTestJSON(discoveryFixture(t, family.name), func(document map[string]any) {
						switch family.name {
						case "space":
							document["sys"].(map[string]any)["organization"].(map[string]any)["sys"].(map[string]any)["id"] = returnedID
						case "environment":
							document["sys"].(map[string]any)["aliasedEnvironment"].(map[string]any)["sys"].(map[string]any)["id"] = returnedID
						case "environment_alias":
							document["environment"].(map[string]any)["sys"].(map[string]any)["id"] = returnedID
						}
					})
					expected := mutateTestJSON(family.expected, func(document map[string]any) {
						switch family.name {
						case "space":
							document["organization_id"] = returnedID
						case "environment":
							document["aliased_environment_id"] = returnedID
						case "environment_alias":
							document["target_environment_id"] = returnedID
						}
					})
					inputs := maps.Clone(family.scopes)
					inputs[family.id] = "item-b"
					singular := discoveryReadTest(t.Context(), t, family.singular, inputs, roundTripFunc(func(request *http.Request) (*http.Response, error) {
						return discoveryHTTPResponse(request, 200, body), nil
					}))
					require.Empty(t, singular.Diagnostics)

					var fields map[string]json.RawMessage
					require.NoError(t, json.Unmarshal([]byte(expected), &fields))

					for name, value := range fields {
						discoveryStateAttribute(t, singular, name, string(value))
					}

					plural := discoveryReadTest(t.Context(), t, family.plural, family.scopes, roundTripFunc(func(request *http.Request) (*http.Response, error) {
						return discoveryHTTPResponse(request, 200, discoveryPage(0, 100, 1, body)), nil
					}))
					require.Empty(t, plural.Diagnostics)
					discoveryStateAttribute(t, plural, family.collection, testJSON([]any{json.RawMessage(expected)}))
				})
			}
		})
	}
}
