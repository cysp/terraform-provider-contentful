package provider

import (
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
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

func TestAppActionParameterValidation(t *testing.T) {
	t.Parallel()

	server, err := providerserver.NewProtocol6WithError(New("test"))()
	require.NoError(t, err)

	for _, test := range []struct {
		name, raw, detail string
		count             int
	}{
		{name: "empty", raw: "[]"},
		{name: "opaque values", raw: `[ {"id":"a","name":"A","type":"Future","default":9007199254740993,"extra":{"nested":null}} ]`},
		{name: "cross-type default", raw: `[{"id":"n","name":"N","type":"Number","default":"42"}]`},
		{name: "empty strings", raw: `[{"id":"","name":"","type":""}]`},
		{name: "null element", raw: `[null]`, detail: "JSON array element 0 must be a parameter definition object.", count: 1},
		{name: "scalar element", raw: `[false]`, detail: "JSON array element 0 must be a parameter definition object.", count: 1},
		{name: "array element", raw: `[[]]`, detail: "JSON array element 0 must be a parameter definition object.", count: 1},
		{name: "empty object", raw: `[{}]`, detail: `JSON array element 0 must contain a string "id" member.`, count: 3},
		{name: "missing id", raw: `[{"name":"A","type":"Symbol"}]`, detail: `JSON array element 0 must contain a string "id" member.`, count: 1},
		{name: "null id", raw: `[{"id":null,"name":"A","type":"Symbol"}]`, detail: `JSON array element 0 must contain a string "id" member.`, count: 1},
		{name: "numeric name", raw: `[{"id":"a","name":42,"type":"Symbol"}]`, detail: `JSON array element 0 must contain a string "name" member.`, count: 1},
		{name: "object type", raw: `[{"id":"a","name":"A","type":{}}]`, detail: `JSON array element 0 must contain a string "type" member.`, count: 1},
		{name: "later element", raw: `[{"id":"a","name":"A","type":"Symbol"},null]`, detail: "JSON array element 1 must be a parameter definition object.", count: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model := appActionTestModel()
			model.ID = types.StringNull()
			model.AppActionID = types.StringNull()
			model.Parameters = jsontypes.NewNormalizedValue(test.raw)
			plan := appActionTestPlan(t, model)
			dynamic, err := tfprotov6.NewDynamicValue(plan.Schema.Type().TerraformType(t.Context()), plan.Raw)
			require.NoError(t, err)
			response, err := server.ValidateResourceConfig(t.Context(), &tfprotov6.ValidateResourceConfigRequest{TypeName: "contentful_app_action", Config: &dynamic})
			require.NoError(t, err)
			require.Len(t, response.Diagnostics, test.count)

			converted, diags := model.ToAppActionData()
			require.Len(t, diags, test.count)

			if test.count == 0 {
				assert.Equal(t, test.raw, string(converted.Parameters))
			} else {
				for _, diagnostic := range response.Diagnostics {
					assert.Equal(t, tftypes.NewAttributePath().WithAttributeName("parameters"), diagnostic.Attribute)
					assert.Equal(t, "Invalid App Action parameters", diagnostic.Summary)
				}

				assert.Equal(t, test.detail, response.Diagnostics[0].Detail)
				assert.Equal(t, test.detail, diags[0].Detail())
			}
		})
	}
}

func TestAppActionParameterValidationUnknownConfig(t *testing.T) {
	t.Parallel()

	server, err := providerserver.NewProtocol6WithError(New("test"))()
	require.NoError(t, err)

	model := appActionTestModel()
	model.ID = types.StringNull()
	model.AppActionID = types.StringNull()
	model.Parameters = jsontypes.NewNormalizedUnknown()
	plan := appActionTestPlan(t, model)
	dynamic, err := tfprotov6.NewDynamicValue(plan.Schema.Type().TerraformType(t.Context()), plan.Raw)
	require.NoError(t, err)
	response, err := server.ValidateResourceConfig(t.Context(), &tfprotov6.ValidateResourceConfigRequest{TypeName: "contentful_app_action", Config: &dynamic})
	require.NoError(t, err)
	require.Empty(t, response.Diagnostics)

	_, diags := model.ToAppActionData()
	require.True(t, diags.HasError())
	assert.Equal(t, []string{"parameters"}, mutationDiagnosticPaths(t, diags))
}

func TestAppActionParameterRequestRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"", "[", "[] []", "invalid json"} {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()

			_, diags := appActionRequestJSON(jsontypes.NewNormalizedValue(raw), path.Root("parameters"), true)
			require.True(t, diags.HasError())
			assert.Equal(t, []string{"parameters"}, mutationDiagnosticPaths(t, diags))
		})
	}
}

func TestAppActionMutationsRejectInvalidParameterPlan(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64

	implementation := appActionTestResource(t, func(http.ResponseWriter, *http.Request) { calls.Add(1) })
	previous := appActionTestModel()
	config := previous
	config.AppActionID = types.StringNull()
	model := previous
	model.Parameters = jsontypes.NewNormalizedValue("[null]")
	plan := appActionTestPlan(t, model)
	state := tfsdk.State(appActionTestPlan(t, previous))
	create := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
	implementation.Create(t.Context(), resource.CreateRequest{Plan: plan, Config: tfsdk.Config(appActionTestPlan(t, config))}, &create)
	require.True(t, create.Diagnostics.HasError())
	assert.Equal(t, "Invalid App Action parameters", create.Diagnostics[0].Summary())

	update := resource.UpdateResponse{State: state}
	implementation.Update(t.Context(), resource.UpdateRequest{Plan: plan, Config: tfsdk.Config(appActionTestPlan(t, config)), State: state}, &update)
	require.True(t, update.Diagnostics.HasError())
	assert.Equal(t, "Invalid App Action parameters", update.Diagnostics[0].Summary())
	assert.True(t, update.State.Raw.Equal(state.Raw))
	assert.Zero(t, calls.Load())
}
