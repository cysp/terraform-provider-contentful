package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests access Read implementations directly to assert that errors never
// publish partial state. Raw HTTP fixtures are independent of CMA constructors.
type discoveryTestFamily struct {
	name         string
	singular     func() datasource.DataSource
	plural       func() datasource.DataSource
	scopes       map[string]any
	id           string
	collection   string
	endpoint     string
	organization string
	expected     string
}

func discoveryTestFamilies() []discoveryTestFamily {
	return []discoveryTestFamily{
		{"space", NewSpaceDataSource, NewSpacesDataSource, map[string]any{}, "space_id", "spaces", "/spaces", "org", `{"space_id":"item-b","name":"Second","organization_id":"org"}`},
		{"environment", NewEnvironmentDataSource, NewEnvironmentsDataSource, map[string]any{"space_id": "space"}, "environment_id", "environments", "/spaces/space/environments", "", `{"environment_id":"item-b","name":"Second","status":"queued","aliased_environment_id":"target"}`},
		{"environment_alias", NewEnvironmentAliasDataSource, NewEnvironmentAliasesDataSource, map[string]any{"space_id": "space"}, "environment_alias_id", "environment_aliases", "/spaces/space/environment_aliases", "", `{"environment_alias_id":"item-b","target_environment_id":"target"}`},
		{"locale", NewLocaleDataSource, NewLocalesDataSource, map[string]any{"space_id": "space", "environment_id": "master"}, "locale_id", "locales", "/spaces/space/environments/master/locales", "", `{"locale_id":"item-b","name":"Second","code":"en-GB","default":true,"fallback_code":null,"optional":false,"content_management_api":true,"content_delivery_api":false}`},
	}
}

func discoveryFixture(t *testing.T, family string) string {
	t.Helper()

	body, err := os.ReadFile("testdata/discovery/" + family + ".json")
	require.NoError(t, err)

	return strings.TrimSpace(string(body))
}

func discoveryReadTest(ctx context.Context, t *testing.T, factory func() datasource.DataSource, inputs map[string]any, transport http.RoundTripper) datasource.ReadResponse {
	t.Helper()

	implementation := factory()

	var schemaResponse datasource.SchemaResponse
	implementation.Schema(ctx, datasource.SchemaRequest{}, &schemaResponse)
	require.False(t, schemaResponse.Diagnostics.HasError(), schemaResponse.Diagnostics)
	schema := schemaResponse.Schema

	values := make(map[string]tftypes.Value, len(schema.Attributes))
	for name, attribute := range schema.Attributes {
		values[name] = tftypes.NewValue(attribute.GetType().TerraformType(ctx), inputs[name])
	}

	client, err := cm.NewClient("https://example.invalid/base", cm.NewAccessTokenSecuritySource("synthetic-token"), cm.WithClient(&http.Client{Transport: transport}))
	require.NoError(t, err)

	configurable, ok := implementation.(datasource.DataSourceWithConfigure)
	require.True(t, ok)

	var configureResponse datasource.ConfigureResponse
	configurable.Configure(ctx, datasource.ConfigureRequest{ProviderData: ContentfulProviderData{client: client}}, &configureResponse)
	require.False(t, configureResponse.Diagnostics.HasError(), configureResponse.Diagnostics)

	response := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	implementation.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: schema, Raw: tftypes.NewValue(schema.Type().TerraformType(ctx), values)}}, &response)

	return response
}

func discoveryHTTPResponse(request *http.Request, status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: request}
}

func discoveryPage(skip, limit, total int, items string) string {
	return fmt.Sprintf(`{"sys":{"type":"Array"},"skip":%d,"limit":%d,"total":%d,"items":[%s]}`, skip, limit, total, items)
}

func discoveryStateAttribute(t *testing.T, response datasource.ReadResponse, name string, expected string) {
	t.Helper()

	var attributes map[string]tftypes.Value
	require.NoError(t, response.State.Raw.As(&attributes))
	actual, exists := attributes[name]
	require.True(t, exists)

	var value any
	require.NoError(t, json.Unmarshal([]byte(expected), &value))
	assert.Equal(t, value, discoveryPlainValue(t, actual), name)
}

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
				discoveryStateAttribute(t, response, "code", `"en-GB"`)
			}

			if family.name == "environment" {
				discoveryStateAttribute(t, response, "status", `"queued"`)
				discoveryStateAttribute(t, response, "aliased_environment_id", `"target"`)
			}
		})
	}
}

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

				second := strings.ReplaceAll(strings.ReplaceAll(body, "item-b", "item-a"), "Second", "First")

				return discoveryHTTPResponse(request, 200, discoveryPage(1, 1, 2, second)), nil
			}))
			require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
			assert.Equal(t, []string{"0", "1"}, offsets)

			first := strings.ReplaceAll(strings.ReplaceAll(family.expected, "item-b", "item-a"), "Second", "First")
			discoveryStateAttribute(t, response, family.collection, "["+first+","+family.expected+"]")
		})
	}
}

func TestDiscoveryDataSourcesPaginationMatchesTeams(t *testing.T) {
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
				{"without metadata", []string{`{"sys":{"type":"Array"},"items":[` + body + `]}`, `{"sys":{"type":"Array"},"items":[]}`}, 1},
				{"empty before total", []string{discoveryPage(0, 100, 10, body), discoveryPage(1, 100, 10, "")}, 1},
				{"increasing total", []string{discoveryPage(0, 100, 2, body), discoveryPage(1, 100, 3, body), discoveryPage(2, 100, 3, body)}, 3},
				{"changing total and duplicates", []string{discoveryPage(0, 100, 3, body), discoveryPage(1, 100, 2, body)}, 2},
				{"echoed metadata does not control progress", []string{discoveryPage(99, 0, 2, body), discoveryPage(99, 0, 1, body)}, 2},
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

					expected := make([]string, testcase.wantCount)
					for i := range expected {
						expected[i] = family.expected
					}

					discoveryStateAttribute(t, response, family.collection, "["+strings.Join(expected, ",")+"]")
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
				{"permission", 403, `{"sys":{"type":"Error","id":"AccessDenied"},"message":"denied"}`, "AccessDenied"},
				{"missing parent", 404, `{"sys":{"type":"Error","id":"NotFound"}}`, "NotFound"},
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

				return discoveryHTTPResponse(request, 200, discoveryPage(0, 100, 0, "")), nil
			}))
			require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
			discoveryStateAttribute(t, response, family.collection, "[]")

			if family.name == "space" {
				discoveryStateAttribute(t, response, "id", `"spaces"`)
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
				{"different ID", 200, strings.ReplaceAll(body, "item-b", "different")},
				{"empty ID", 200, strings.ReplaceAll(body, "item-b", "")},
				{"not found", 404, `{"sys":{"type":"Error","id":"NotFound"}}`},
				{"unauthorized", 401, `{"sys":{"type":"Error","id":"AccessTokenInvalid"}}`},
				{"forbidden", 403, `{"sys":{"type":"Error","id":"AccessDenied"}}`},
				{"wrong type", 200, `{"sys":{"type":"Array"},"items":[]}`},
				{"null body", 200, `null`},
				{"trailing body", 200, body + ` {}`},
			}
			if family.name != "space" {
				cases = append(cases, struct {
					name   string
					status int
					body   string
				}{"different space", 200, strings.ReplaceAll(body, `"id":"space"`, `"id":"other-space"`)})
			}

			if family.name == "locale" {
				cases = append(cases, struct {
					name   string
					status int
					body   string
				}{"different environment", 200, strings.ReplaceAll(body, `"id":"master"`, `"id":"concrete"`)})
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

func TestLocaleDataSourceFallbackValues(t *testing.T) {
	t.Parallel()

	body := discoveryFixture(t, "locale")
	for _, testcase := range []struct {
		name        string
		replacement string
		expected    string
	}{
		{"absent", "", `null`}, {"null", `"fallbackCode":null,`, `null`}, {"empty", `"fallbackCode":"",`, `""`}, {"code", `"fallbackCode":"fr-FR",`, `"fr-FR"`},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			response := discoveryReadTest(t.Context(), t, NewLocaleDataSource, map[string]any{"space_id": "space", "environment_id": "master", "locale_id": "item-b"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
				return discoveryHTTPResponse(request, 200, strings.ReplaceAll(body, `"fallbackCode":null,`, testcase.replacement)), nil
			}))
			require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
			discoveryStateAttribute(t, response, "fallback_code", testcase.expected)
		})
	}
}

func TestEnvironmentDataSourceStatusesAndSharedProjection(t *testing.T) {
	t.Parallel()

	body := discoveryFixture(t, "environment")
	for _, status := range []string{"queued", "inProgress", "ready", "failed", "future-status"} {
		t.Run(status, func(t *testing.T) {
			t.Parallel()

			count := 0
			response := discoveryReadTest(t.Context(), t, NewEnvironmentDataSource, map[string]any{"space_id": "space", "environment_id": "item-b"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
				count++

				return discoveryHTTPResponse(request, 200, strings.ReplaceAll(body, "queued", status)), nil
			}))
			require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
			assert.Equal(t, 1, count)
			discoveryStateAttribute(t, response, "status", fmt.Sprintf("%q", status))
		})
	}

	for _, aliased := range []bool{false, true} {
		t.Run(strconv.FormatBool(aliased), func(t *testing.T) {
			t.Parallel()

			responseBody := body
			if !aliased {
				responseBody = strings.ReplaceAll(body, `,"aliasedEnvironment":{"sys":{"id":"target","type":"Link","linkType":"Environment"}}`, "")
			}

			client, err := cm.NewClient("https://example.invalid", cm.NewAccessTokenSecuritySource("synthetic"), cm.WithClient(&http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				return discoveryHTTPResponse(request, 200, responseBody), nil
			})}))
			require.NoError(t, err)
			result, err := client.GetEnvironment(t.Context(), cm.GetEnvironmentParams{SpaceID: "space", EnvironmentID: "item-b"})
			require.NoError(t, err)

			entity, ok := result.(*cm.Environment)
			require.True(t, ok)

			resource := NewEnvironmentResourceModelFromResponse(*entity)
			waiter := NewEnvironmentStatusReadyModelFromResponse(*entity)

			assert.Equal(t, "space/item-b", resource.ID.ValueString())
			assert.Equal(t, "Second", resource.Name.ValueString())
			assert.Equal(t, "queued", resource.Status.ValueString())
			assert.Equal(t, "space/item-b", waiter.ID.ValueString())
			assert.Equal(t, "queued", waiter.Status.ValueString())
			assert.Equal(t, 7, entity.Sys.Version)
		})
	}
}

func TestEnvironmentDataSourceReducedPublishedExampleFails(t *testing.T) {
	t.Parallel()

	for _, body := range []string{
		`{"sys":{"id":"item-b","type":"Environment","version":1,"space":{"sys":{"type":"Link","linkType":"Space","id":"space"}}}}`,
		`{"sys":{"id":"item-b","type":"Environment","version":1,"space":{"sys":{"type":"Link","linkType":"Space","id":"space"}},"status":{"sys":{"type":"Link","linkType":"Status","id":"ready"}}}}`,
	} {
		response := discoveryReadTest(t.Context(), t, NewEnvironmentDataSource, map[string]any{"space_id": "space", "environment_id": "item-b"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
			return discoveryHTTPResponse(request, 200, body), nil
		}))
		assert.True(t, response.Diagnostics.HasError())
		assert.True(t, response.State.Raw.IsNull())
	}
}

func TestEnvironmentAliasDataSourceBothTypeSpellings(t *testing.T) {
	t.Parallel()

	body := discoveryFixture(t, "environment_alias")
	for _, spelling := range []string{"Environment Alias", "EnvironmentAlias"} {
		t.Run(spelling, func(t *testing.T) {
			t.Parallel()

			responseBody := strings.ReplaceAll(body, "Environment Alias", spelling)
			client, err := cm.NewClient("https://example.invalid", cm.NewAccessTokenSecuritySource("synthetic"), cm.WithClient(&http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				return discoveryHTTPResponse(request, 200, responseBody), nil
			})}))
			require.NoError(t, err)
			result, err := client.GetEnvironmentAlias(t.Context(), cm.GetEnvironmentAliasParams{SpaceID: "space", EnvironmentAliasID: "item-b"})
			require.NoError(t, err)

			entity, ok := result.(*cm.EnvironmentAlias)
			require.True(t, ok)

			model := NewEnvironmentAliasResourceModelFromResponse(*entity)
			assert.Equal(t, "space/item-b", model.ID.ValueString())
			assert.Equal(t, "item-b", model.EnvironmentAliasID.ValueString())
			assert.Equal(t, "target", model.TargetEnvironmentID.ValueString())
			assert.Equal(t, 9, entity.Sys.Version)
		})
	}
}

func TestSpacesDataSourceOrganizationScopeContradiction(t *testing.T) {
	t.Parallel()
	body := discoveryFixture(t, "space")
	response := discoveryReadTest(t.Context(), t, NewSpacesDataSource, map[string]any{"organization_id": "another-org"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return discoveryHTTPResponse(request, 200, discoveryPage(0, 100, 1, body)), nil
	}))
	assert.True(t, response.Diagnostics.HasError())
	assert.True(t, response.State.Raw.IsNull())
}

func TestDiscoveryDataSourcesEscapedIDs(t *testing.T) {
	t.Parallel()

	for _, family := range discoveryTestFamilies() {
		t.Run(family.name, func(t *testing.T) {
			t.Parallel()
			body := strings.ReplaceAll(discoveryFixture(t, family.name), "item-b", "a%2Fb?c#d")

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

func TestLocalesDataSourceAliasRetargetingHasNoSnapshotGuarantee(t *testing.T) {
	t.Parallel()
	body := discoveryFixture(t, "locale")
	requests := 0
	// Two different targets could return distinct Locales under the same alias
	// metadata. This fixture demonstrates the observation limit, not CMA timing.
	response := discoveryReadTest(t.Context(), t, NewLocalesDataSource, map[string]any{"space_id": "space", "environment_id": "master"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++

		assert.Equal(t, "/base/spaces/space/environments/master/locales", request.URL.Path)

		if requests == 1 {
			return discoveryHTTPResponse(request, 200, discoveryPage(0, 1, 2, body)), nil
		}

		return discoveryHTTPResponse(request, 200, discoveryPage(1, 1, 2, strings.ReplaceAll(strings.ReplaceAll(body, "item-b", "other-target-locale"), "en-GB", "fr-FR"))), nil
	}))
	require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
	assert.Equal(t, 2, requests)

	var data LocalesDataSourceModel
	require.False(t, response.State.Get(t.Context(), &data).HasError())
	require.Len(t, data.Locales, 2)
	assert.Equal(t, "master", data.EnvironmentID.ValueString())
	assert.Equal(t, "en-GB", data.Locales[0].Code.ValueString())
	assert.Equal(t, "fr-FR", data.Locales[1].Code.ValueString())
}

func discoveryPlainValue(t *testing.T, value tftypes.Value) any {
	t.Helper()

	if value.IsNull() {
		return nil
	}

	switch value.Type().(type) {
	case tftypes.Object:
		var fields map[string]tftypes.Value
		require.NoError(t, value.As(&fields))

		result := make(map[string]any, len(fields))
		for name, field := range fields {
			result[name] = discoveryPlainValue(t, field)
		}

		return result
	case tftypes.List:
		var items []tftypes.Value
		require.NoError(t, value.As(&items))

		result := make([]any, 0, len(items))
		for _, item := range items {
			result = append(result, discoveryPlainValue(t, item))
		}

		return result
	default:
		if value.Type().Is(tftypes.String) {
			var result string
			require.NoError(t, value.As(&result))

			return result
		}

		var result bool
		require.NoError(t, value.As(&result))

		return result
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

			second := strings.ReplaceAll(body, "item-b", "item-a")
			second = strings.ReplaceAll(strings.ReplaceAll(second, `"id":"org"`, `"id":"other"`), `"id":"space"`, `"id":"other"`)
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

func TestDiscoveryCollectionsPreserveIrregularItemIDs(t *testing.T) {
	t.Parallel()

	for _, family := range discoveryTestFamilies() {
		t.Run(family.name, func(t *testing.T) {
			t.Parallel()

			for _, returnedID := range []string{"", ".", "..", "a/b"} {
				t.Run(returnedID, func(t *testing.T) {
					t.Parallel()
					body := discoveryFixture(t, family.name)
					irregular := strings.ReplaceAll(body, "item-b", returnedID)
					count := 0
					response := discoveryReadTest(t.Context(), t, family.plural, family.scopes, roundTripFunc(func(request *http.Request) (*http.Response, error) {
						count++
						if count == 1 {
							assert.Equal(t, "0", request.URL.Query().Get("skip"))

							return discoveryHTTPResponse(request, 200, discoveryPage(0, 100, 3, body)), nil
						}

						require.Equal(t, 2, count)
						assert.Equal(t, "1", request.URL.Query().Get("skip"))

						return discoveryHTTPResponse(request, 200, discoveryPage(1, 100, 3, irregular+","+irregular)), nil
					}))
					require.Empty(t, response.Diagnostics)
					assert.Equal(t, 2, count)

					expected := strings.ReplaceAll(family.expected, "item-b", returnedID)
					discoveryStateAttribute(t, response, family.collection, "["+expected+","+expected+","+family.expected+"]")

					var attributes map[string]tftypes.Value
					require.NoError(t, response.State.Raw.As(&attributes))

					var items []tftypes.Value
					require.NoError(t, attributes[family.collection].As(&items))

					var item map[string]tftypes.Value
					require.NoError(t, items[0].As(&item))

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

func TestDiscoveryDataSourcesPreserveComputedReferences(t *testing.T) {
	t.Parallel()

	for _, family := range discoveryTestFamilies() {
		if family.name == "locale" {
			continue
		}

		t.Run(family.name, func(t *testing.T) {
			t.Parallel()

			reference := "target"
			if family.name == "space" {
				reference = "org"
			}

			for _, returnedID := range []string{"", ".", "..", "a/b"} {
				t.Run(returnedID, func(t *testing.T) {
					t.Parallel()
					body := strings.ReplaceAll(discoveryFixture(t, family.name), `"id":"`+reference+`"`, `"id":"`+returnedID+`"`)
					expected := strings.ReplaceAll(family.expected, `"`+reference+`"`, `"`+returnedID+`"`)
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
					discoveryStateAttribute(t, plural, family.collection, "["+expected+"]")
				})
			}
		})
	}
}
