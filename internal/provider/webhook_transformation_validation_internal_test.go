package provider

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func webhookTransformationValidationPlan(t *testing.T, transformation map[string]any) tfsdk.Plan {
	t.Helper()
	schema := WebhookResourceSchema(t.Context())
	raw, err := json.Marshal(map[string]any{"space_id": "space", "name": "Webhook", "url": "https://example.invalid/{ /payload/sys/id }", "topics": []string{"Entry.publish"}, "active": true, "transformation": transformation})
	require.NoError(t, err)
	value, err := tftypes.ValueFromJSONWithOpts(raw, schema.Type().TerraformType(t.Context()), tftypes.ValueFromJSONOpts{})
	require.NoError(t, err)

	return tfsdk.Plan{Schema: schema, Raw: value}
}

func TestWebhookTransformationValidation(t *testing.T) {
	t.Parallel()

	server, err := providerserver.NewProtocol6WithError(New("test"))()
	require.NoError(t, err)

	for _, test := range []struct {
		name, attribute, value string
		invalid                bool
	}{
		{"POST", "method", "POST", false},
		{"GET", "method", "GET", false},
		{"PUT", "method", "PUT", false},
		{"PATCH", "method", "PATCH", false},
		{"DELETE", "method", "DELETE", false},
		{"unsupported method", "method", "HEAD", true},
		{"lowercase method", "method", "post", true},
		{"empty method", "method", "", true},
		{"management JSON", "content_type", "application/vnd.contentful.management.v1+json", false},
		{"management charset", "content_type", "application/vnd.contentful.management.v1+json; charset=utf-8", false},
		{"JSON", "content_type", "application/json", false},
		{"JSON charset", "content_type", "application/json; charset=utf-8", false},
		{"form", "content_type", "application/x-www-form-urlencoded", false},
		{"form charset", "content_type", "application/x-www-form-urlencoded; charset=utf-8", false},
		{"unsupported content type", "content_type", "text/plain", true},
		{"empty content type", "content_type", "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			plan := webhookTransformationValidationPlan(t, map[string]any{test.attribute: test.value, "body": "[false,42,null]"})
			dynamic, err := tfprotov6.NewDynamicValue(plan.Schema.Type().TerraformType(t.Context()), plan.Raw)
			require.NoError(t, err)
			response, err := server.ValidateResourceConfig(t.Context(), &tfprotov6.ValidateResourceConfigRequest{TypeName: "contentful_webhook", Config: &dynamic})
			require.NoError(t, err)

			var model WebhookModel
			require.Empty(t, plan.Get(t.Context(), &model))
			converted, diags := ToOptNilWebhookDefinitionDataTransformation(path.Root("transformation"), model.Transformation)

			if test.invalid {
				require.Len(t, response.Diagnostics, 1)
				assert.Equal(t, tftypes.NewAttributePath().WithAttributeName("transformation").WithAttributeName(test.attribute), response.Diagnostics[0].Attribute)
				assert.Equal(t, "Invalid webhook transformation value", response.Diagnostics[0].Summary)
				require.Len(t, diags, 1)
				assert.False(t, converted.IsSet())
			} else {
				require.Empty(t, response.Diagnostics)
				require.Empty(t, diags)
				assert.Equal(t, "[false,42,null]", string(converted.Value.Body))

				if test.attribute == "method" {
					assert.Equal(t, test.value, converted.Value.Method.Value)
				} else {
					assert.Equal(t, test.value, converted.Value.ContentType.Value)
				}
			}
		})
	}

	for _, value := range []types.String{types.StringNull(), types.StringUnknown()} {
		plan := webhookTransformationValidationPlan(t, map[string]any{})
		require.Empty(t, plan.SetAttribute(t.Context(), path.Root("transformation").AtName("method"), value))
		dynamic, err := tfprotov6.NewDynamicValue(plan.Schema.Type().TerraformType(t.Context()), plan.Raw)
		require.NoError(t, err)
		response, err := server.ValidateResourceConfig(t.Context(), &tfprotov6.ValidateResourceConfigRequest{TypeName: "contentful_webhook", Config: &dynamic})
		require.NoError(t, err)
		require.Empty(t, response.Diagnostics)
	}
}

func TestWebhookTransformationInvalidPlanStopsWrites(t *testing.T) {
	t.Parallel()

	for _, attribute := range []string{"method", "content_type"} {
		t.Run(attribute, func(t *testing.T) {
			t.Parallel()
			plan := webhookTransformationValidationPlan(t, map[string]any{attribute: "unsupported"})
			config := tfsdk.Config(webhookTransformationValidationPlan(t, map[string]any{}))
			client, count := mutationRequestCountingClient(t)
			implementation := webhookResource{providerData: ContentfulProviderData{client: client}}
			create := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
			implementation.Create(t.Context(), resource.CreateRequest{Plan: plan, Config: config}, &create)
			require.Equal(t, []string{"transformation." + attribute}, mutationDiagnosticPaths(t, create.Diagnostics))

			update := resource.UpdateResponse{State: tfsdk.State{Schema: plan.Schema}}
			implementation.Update(t.Context(), resource.UpdateRequest{Plan: plan, Config: config}, &update)
			require.Equal(t, []string{"transformation." + attribute}, mutationDiagnosticPaths(t, update.Diagnostics))
			require.Zero(t, count.Load())
		})
	}
}
