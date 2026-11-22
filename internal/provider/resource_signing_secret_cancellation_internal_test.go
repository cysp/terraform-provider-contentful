package provider

import (
	"context"
	"maps"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"testing/synctest"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest // Own the one process-wide workspace while exercising independent resource instances.
func TestSigningSecretCreatesShareCancellation(t *testing.T) {
	var requests atomic.Int64

	client, err := cm.NewClient("https://contentful.invalid", cm.NewAccessTokenSecuritySource("test"), cm.WithClient(&http.Client{
		Transport: contentfulRetryTestRoundTripper(func(req *http.Request) (*http.Response, error) {
			requests.Add(1)

			return contentfulRetryTestResponse(req, http.StatusInternalServerError), nil
		}),
	}))
	require.NoError(t, err)

	providerData := ContentfulProviderData{client: client}
	signingValue := strings.Repeat("s", 64)
	cases := []struct {
		name           string
		implementation resource.Resource
		schema         schema.Schema
		planned        map[string]any
	}{
		{
			name:           "app signing",
			implementation: &appSigningSecretResource{providerData: providerData},
			schema:         AppSigningSecretResourceSchema(t.Context()),
			planned:        map[string]any{"organization_id": "org", "app_definition_id": "app"},
		},
		{
			name:           "webhook signing",
			implementation: &webhookSigningSecretResource{providerData: providerData},
			schema:         WebhookSigningSecretResourceSchema(t.Context()),
			planned:        map[string]any{"space_id": "space"},
		},
	}

	require.True(t, signingSecretHashSlot.TryAcquire(1))

	defer signingSecretHashSlot.Release(1)

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		results := make(chan resource.CreateResponse, len(cases))
		for _, testCase := range cases {
			typ := testCase.schema.Type().TerraformType(ctx)
			plan, err := tftypes.ValueFromJSONWithOpts([]byte(testJSON(testCase.planned)), typ, tftypes.ValueFromJSONOpts{})
			require.NoError(t, err, testCase.name)

			configured := maps.Clone(testCase.planned)
			configured["value_wo"] = signingValue
			config, err := tftypes.ValueFromJSONWithOpts([]byte(testJSON(configured)), typ, tftypes.ValueFromJSONOpts{})
			require.NoError(t, err, testCase.name)

			go func() {
				resp := resource.CreateResponse{State: tfsdk.State{Schema: testCase.schema, Raw: tftypes.NewValue(typ, nil)}}
				testCase.implementation.Create(ctx, resource.CreateRequest{
					Plan: tfsdk.Plan{Schema: testCase.schema, Raw: plan}, Config: tfsdk.Config{Schema: testCase.schema, Raw: config},
				}, &resp)

				results <- resp
			}()
		}

		synctest.Wait()
		assert.Empty(t, results, "both resources must wait for the process slot")
		cancel()

		for range cases {
			resp := <-results
			require.Len(t, resp.Diagnostics.Errors(), 1)
			assert.Equal(t, "Write-only secret operation cancelled", resp.Diagnostics.Errors()[0].Summary())
			assert.Equal(t, context.Canceled.Error(), resp.Diagnostics.Errors()[0].Detail())
			assert.True(t, resp.State.Raw.IsNull())
		}

		assert.Zero(t, requests.Load())
	})
}
