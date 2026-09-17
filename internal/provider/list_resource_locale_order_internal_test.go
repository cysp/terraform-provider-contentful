package provider

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocaleListResourceStreamsAPIOrder(t *testing.T) {
	t.Parallel()

	var (
		requests []url.Values
		emitted  int
	)

	client, err := cm.NewClient("https://example.invalid", cm.NewAccessTokenSecuritySource("synthetic-token"), cm.WithClient(&http.Client{
		Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodGet, request.Method)
			assert.Equal(t, "/spaces/space/environments/environment/locales", request.URL.Path)
			requests = append(requests, request.URL.Query())

			status := http.StatusOK

			var body string

			switch request.URL.Query().Get("skip") {
			case "0":
				body = `{"sys":{"type":"Array"},"total":3,"skip":0,"limit":1,"items":[{"sys":{"id":"z","type":"Locale","space":{"sys":{"id":"space","type":"Link","linkType":"Space"}},"environment":{"sys":{"id":"environment","type":"Link","linkType":"Environment"}}},"name":"First by name","code":"z-code","fallbackCode":null,"default":false,"optional":false,"contentDeliveryApi":true,"contentManagementApi":true}]}`
			case "1":
				assert.Equal(t, 1, emitted, "the first page must be emitted before requesting the second")

				body = `{"sys":{"type":"Array"},"total":3,"skip":1,"limit":1,"items":[{"sys":{"id":"a","type":"Locale","space":{"sys":{"id":"space","type":"Link","linkType":"Space"}},"environment":{"sys":{"id":"environment","type":"Link","linkType":"Environment"}}},"name":"Second by name","code":"a-code","fallbackCode":null,"default":false,"optional":false,"contentDeliveryApi":true,"contentManagementApi":true}]}`

			default:
				t.Errorf("unexpected page request: %s", request.URL)

				status = http.StatusBadRequest
				body = `{"sys":{"type":"Error","id":"BadRequest"},"message":"unexpected page"}`
			}

			return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
		}),
	}))
	require.NoError(t, err)

	ctx := t.Context()
	schema := LocaleListResourceConfigSchema(ctx)
	config := tfsdk.Config{
		Schema: schema,
		Raw: tftypes.NewValue(schema.Type().TerraformType(ctx), map[string]tftypes.Value{
			"space_id":       tftypes.NewValue(tftypes.String, "space"),
			"environment_id": tftypes.NewValue(tftypes.String, "environment"),
		}),
	}

	var identitySchema resource.IdentitySchemaResponse
	(&localeResource{}).IdentitySchema(ctx, resource.IdentitySchemaRequest{}, &identitySchema)

	implementation := localeResource{providerData: ContentfulProviderData{client: client}}

	var stream list.ListResultsStream
	implementation.List(ctx, list.ListRequest{
		Config: config, Limit: 2, IncludeResource: true,
		ResourceSchema: LocaleResourceSchema(ctx), ResourceIdentitySchema: identitySchema.IdentitySchema,
	}, &stream)

	var results []list.ListResult
	for result := range stream.Results {
		results = append(results, result)
		emitted++
	}

	assert.Equal(t, []url.Values{
		{"skip": {"0"}, "limit": {"2"}},
		{"skip": {"1"}, "limit": {"1"}},
	}, requests)

	ids := make([]string, 0, len(results))

	for _, result := range results {
		require.False(t, result.Diagnostics.HasError(), result.Diagnostics)

		var identity LocaleIdentityModel
		require.Empty(t, result.Identity.Get(ctx, &identity))
		ids = append(ids, identity.LocaleID.ValueString())

		var model LocaleModel
		require.Empty(t, result.Resource.Get(ctx, &model))
		assert.Equal(t, identity.LocaleID.ValueString()+"-code", model.Code.ValueString())
	}

	assert.Equal(t, []string{"z", "a"}, ids)
}
