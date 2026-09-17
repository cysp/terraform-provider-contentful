package provider

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errListFetchUnavailable = errors.New("list transport unavailable")

func TestListResourceFetchDiagnostics(t *testing.T) {
	t.Parallel()

	for _, family := range []struct {
		name    string
		path    string
		summary string
		create  func(*cm.Client) list.ListResource
	}{
		{"entries", "/spaces/space/environments/environment/entries", "Failed to list entries", func(client *cm.Client) list.ListResource {
			return &entryListResource{providerData: ContentfulProviderData{client: client}}
		}},
		{"content types", "/spaces/space/environments/environment/content_types", "Failed to list content types", func(client *cm.Client) list.ListResource {
			return &contentTypeListResource{providerData: ContentfulProviderData{client: client}}
		}},
		{"locales", "/spaces/space/environments/environment/locales", "Failed to list locales", func(client *cm.Client) list.ListResource {
			return &localeResource{providerData: ContentfulProviderData{client: client}}
		}},
	} {
		t.Run(family.name, func(t *testing.T) {
			t.Parallel()

			for _, testcase := range []struct {
				name   string
				status int
				body   string
				detail string
			}{
				{"transport", 0, "", "list transport unavailable"},
				{"not found", 404, `{"sys":{"type":"Error","id":"NotFound"},"message":"Environment not found"}`, "Error: NotFound: Environment not found"},
				{"forbidden", 403, `{"sys":{"type":"Error","id":"AccessDenied"},"message":"denied"}`, "Error: AccessDenied: denied"},
				{"malformed", 200, `{"items":`, "decode"},
			} {
				t.Run(testcase.name, func(t *testing.T) {
					t.Parallel()

					count := 0
					client, err := cm.NewClient("https://example.invalid", cm.NewAccessTokenSecuritySource("synthetic-token"), cm.WithClient(&http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
						count++

						assert.Equal(t, http.MethodGet, request.Method)
						assert.Equal(t, family.path, request.URL.Path)
						assert.Equal(t, "0", request.URL.Query().Get("skip"))
						assert.Equal(t, "100", request.URL.Query().Get("limit"))

						if testcase.status == 0 {
							return nil, errListFetchUnavailable
						}

						return &http.Response{StatusCode: testcase.status, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(testcase.body)), Request: request}, nil
					})}))
					require.NoError(t, err)

					implementation := family.create(client)

					var schemaResponse list.ListResourceSchemaResponse
					implementation.ListResourceConfigSchema(t.Context(), list.ListResourceSchemaRequest{}, &schemaResponse)

					values := make(map[string]tftypes.Value, len(schemaResponse.Schema.Attributes))
					for name, attribute := range schemaResponse.Schema.Attributes {
						var value any

						switch name {
						case "space_id":
							value = "space"
						case "environment_id":
							value = "environment"
						}

						values[name] = tftypes.NewValue(attribute.GetType().TerraformType(t.Context()), value)
					}

					config := tfsdk.Config{Schema: schemaResponse.Schema, Raw: tftypes.NewValue(schemaResponse.Schema.Type().TerraformType(t.Context()), values)}

					var stream list.ListResultsStream
					implementation.List(t.Context(), list.ListRequest{Config: config}, &stream)
					result := requireSingleDiagnosticOnlyListResult(t, stream)
					require.Len(t, result.Diagnostics, 1)
					assert.Equal(t, family.summary, result.Diagnostics[0].Summary())
					assert.Contains(t, result.Diagnostics[0].Detail(), testcase.detail)
					assert.Equal(t, 1, count)
				})
			}
		})
	}
}
