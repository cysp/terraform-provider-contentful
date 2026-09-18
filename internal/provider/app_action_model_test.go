//nolint:testpackage // Exercise effective Plan conversion and response reconciliation directly.
package provider

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func appActionTestModel() AppActionModel {
	return AppActionModel{AppActionBaseModel: AppActionBaseModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID("org", "app", "action"), OrganizationID: types.StringValue("org"), AppDefinitionID: types.StringValue("app"),
		AppActionFields: AppActionFields{AppActionID: types.StringValue("action"), Name: types.StringValue("Action"), Category: types.StringValue("Custom"), Type: types.StringValue("endpoint"), URL: types.StringValue("https://example.invalid/action"), Parameters: jsontypes.NewNormalizedValue(`[]`)},
	}, Timeouts: TimeoutsNull()}
}

func appActionTestResponse() cm.AppAction {
	return cm.AppAction{Sys: cm.AppActionSys{ID: "action", Type: cm.AppActionSysTypeAppAction, Organization: cm.NewOrganizationLink("org"), AppDefinition: cm.NewAppDefinitionLink("app")}, Name: "Action", Category: "Custom", Type: "endpoint", URL: cm.NewOptString("https://example.invalid/action"), Parameters: jx.Raw(`[]`)}
}

func TestAppActionRequestOwnership(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name      string
		alter     func(*AppActionModel, *AppActionModel)
		want      string
		wantError bool
	}{
		{name: "legacy empty", want: "[]"},
		{name: "known plan config null", alter: func(_ *AppActionModel, config *AppActionModel) { config.Parameters = jsontypes.NewNormalizedNull() }, want: "[]"},
		{name: "schema unknown unconfigured legacy", alter: func(plan, config *AppActionModel) {
			config.Parameters = jsontypes.NewNormalizedNull()
			plan.Parameters = jsontypes.NewNormalizedUnknown()
			plan.ParametersSchema = jsontypes.NewNormalizedValue(`{}`)
		}},
		{name: "unknown configured legacy", alter: func(plan, _ *AppActionModel) { plan.Parameters = jsontypes.NewNormalizedUnknown() }, wantError: true},
		{name: "missing custom input", alter: func(plan, _ *AppActionModel) { plan.Parameters = jsontypes.NewNormalizedNull() }, wantError: true},
		{name: "both custom inputs", alter: func(plan, _ *AppActionModel) { plan.ParametersSchema = jsontypes.NewNormalizedValue(`{}`) }, wantError: true},
		{name: "builtin known defaults omitted", alter: func(plan, config *AppActionModel) {
			plan.Category = types.StringValue("Entries.v1.0")
			config.Parameters = jsontypes.NewNormalizedNull()
		}},
		{name: "effective builtin category", alter: func(plan, config *AppActionModel) {
			plan.Category = types.StringValue("Entries.v1.0")
			config.Category = types.StringValue("Custom")
			config.Parameters = jsontypes.NewNormalizedNull()
		}},
		{name: "deferred builtin parameters", alter: func(plan, config *AppActionModel) {
			plan.Category = types.StringValue("Entries.v1.0")
			config.Category = plan.Category
			config.Parameters = jsontypes.NewNormalizedUnknown()
		}, wantError: true},
		{name: "unsupported category", alter: func(plan, _ *AppActionModel) { plan.Category = types.StringValue("Future.v1.0") }, wantError: true},
		{name: "unknown name", alter: func(plan, _ *AppActionModel) { plan.Name = types.StringUnknown() }, wantError: true},
		{name: "unknown description", alter: func(plan, _ *AppActionModel) { plan.Description = types.StringUnknown() }, wantError: true},
		{name: "both executors", alter: func(plan, _ *AppActionModel) { plan.FunctionID = types.StringValue("fn") }, wantError: true},
		{name: "HTTP endpoint", alter: func(plan, _ *AppActionModel) { plan.URL = types.StringValue("http://example.invalid") }, wantError: true},
		{name: "service validates legacy members", alter: func(plan, _ *AppActionModel) { plan.Parameters = jsontypes.NewNormalizedValue(`[null]`) }, want: `[null]`},
		{name: "JSON null schema", alter: func(plan, _ *AppActionModel) { plan.ResultSchema = jsontypes.NewNormalizedValue(`null`) }, wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			plan := appActionTestModel()

			config := plan
			if test.alter != nil {
				test.alter(&plan, &config)
			}

			body, diags := plan.ToAppActionData(config)
			require.Equal(t, test.wantError, diags.HasError(), "%v", diags)

			if !test.wantError {
				assert.Equal(t, test.want, string(body.Parameters))
			}
		})
	}
}

func TestAppActionResponseReconciliation(t *testing.T) {
	t.Parallel()

	plan := appActionTestModel()
	plan.ParametersSchema = jsontypes.NewNormalizedValue(`{ "type": "object", "properties": {} }`)
	plan.Parameters = jsontypes.NewNormalizedNull()
	response := appActionTestResponse()
	response.Parameters = nil
	response.ParametersSchema = jx.Raw(`{"properties":{},"type":"object"}`)
	data, consistency := reconcileAppActionResponse(t.Context(), response, plan, plan.AppActionBaseModel)
	require.False(t, consistency.HasError(), "%v", consistency)
	assert.Equal(t, plan.ParametersSchema, data.ParametersSchema)

	response.Description = cm.NewOptString("unexpected")
	data, consistency = reconcileAppActionResponse(t.Context(), response, plan, plan.AppActionBaseModel)
	require.True(t, consistency.HasError())
	assert.Equal(t, "unexpected", data.Description.ValueString())
	assert.Equal(t, jsontypes.NewNormalizedValue(`{"properties":{},"type":"object"}`), data.ParametersSchema, "no partial planned JSON overlay after mismatch")

	response.Sys.ID = "other"
	response.Sys.Organization = cm.NewOrganizationLink("other-org")
	data, consistency = reconcileAppActionResponse(t.Context(), response, plan, plan.AppActionBaseModel)
	require.True(t, consistency.HasError())
	assert.Equal(t, "org/app/action", data.ID.ValueString())
}

func TestAppActionBuiltInResponseAndConstraint(t *testing.T) {
	t.Parallel()

	plan := appActionTestModel()
	plan.Category = types.StringValue("Entries.v1.0")
	plan.ParametersSchema = jsontypes.NewNormalizedValue(`{}`)
	response := appActionTestResponse()
	response.Category = "Entries.v1.0"
	response.ParametersSchema = jx.Raw(`{}`)
	response.Parameters = jx.Raw(`[{"id":"entryIds"}]`)
	data, consistency := reconcileAppActionResponse(t.Context(), response, plan, plan.AppActionBaseModel)
	require.True(t, consistency.HasError())
	assert.JSONEq(t, `[{"id":"entryIds"}]`, data.Parameters.ValueString())
	assert.Equal(t, `{}`, data.ParametersSchema.ValueString())

	plan.Parameters = jsontypes.NewNormalizedUnknown()
	_, consistency = reconcileAppActionResponse(t.Context(), response, plan, plan.AppActionBaseModel)
	require.False(t, consistency.HasError(), "%v", consistency)
}

func TestAppActionResponsePreservesIrregularValues(t *testing.T) {
	t.Parallel()

	response := appActionTestResponse()
	response.Type = "future"
	response.Category = "Future.v1.0"
	response.Function = cm.NewOptFunctionLink(cm.NewFunctionLink("fn"))
	response.ParametersSchema = jx.Raw(`null`)
	data, identityDiags := appActionResponse(response, appActionTestModel().AppActionBaseModel)
	require.False(t, identityDiags.HasError())
	assert.Equal(t, "future", data.Type.ValueString())
	assert.Equal(t, "Future.v1.0", data.Category.ValueString())
	assert.Equal(t, "fn", data.FunctionID.ValueString())
	assert.False(t, data.URL.IsNull())
	assert.Equal(t, "null", data.ParametersSchema.ValueString())
}
