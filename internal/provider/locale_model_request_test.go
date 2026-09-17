package provider_test

import (
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validLocaleRequestModel() provider.LocaleModel {
	return provider.LocaleModel{
		LocaleIdentityModel: provider.LocaleIdentityModel{
			SpaceID:       types.StringValue("space"),
			EnvironmentID: types.StringValue("environment"),
			LocaleID:      types.StringValue("locale"),
		},
		Name:                 types.StringValue("English (Australia)"),
		Code:                 types.StringValue("en-AU"),
		FallbackCode:         types.StringNull(),
		ContentDeliveryAPI:   types.BoolValue(true),
		ContentManagementAPI: types.BoolValue(true),
		Optional:             types.BoolValue(false),
	}
}

func TestLocaleResourceRejectsUnresolvedValuesBeforeIO(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		text     types.String
		boolean  types.Bool
		identity bool
		paths    []string
	}{
		"unknown payload": {text: types.StringUnknown(), boolean: types.BoolUnknown(), paths: []string{"name", "code", "fallback_code", "content_delivery_api", "content_management_api", "optional"}},
		"null payload":    {text: types.StringNull(), boolean: types.BoolNull(), paths: []string{"name", "code", "content_delivery_api", "content_management_api", "optional"}},
		"identity":        {identity: true, paths: []string{"space_id", "environment_id"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := validLocaleRequestModel()

			model.Timeouts = provider.TimeoutsNull()
			if test.identity {
				model.SpaceID, model.EnvironmentID, model.LocaleID = types.StringUnknown(), types.StringNull(), types.StringUnknown()
			} else {
				model.Name, model.Code, model.FallbackCode = test.text, test.text, test.text
				model.ContentDeliveryAPI, model.ContentManagementAPI, model.Optional = test.boolean, test.boolean, test.boolean
			}

			plan := tfsdk.Plan{Schema: provider.LocaleResourceSchema(t.Context())}
			require.Empty(t, plan.Set(t.Context(), model))

			// An unconfigured resource has no client: validation must stop before I/O.
			implementation := provider.NewLocaleResource()
			create := resource.CreateResponse{}
			implementation.Create(t.Context(), resource.CreateRequest{Plan: plan}, &create)
			assert.Equal(t, test.paths, attributeDiagnosticPaths(t, create.Diagnostics))

			update := resource.UpdateResponse{}
			implementation.Update(t.Context(), resource.UpdateRequest{Plan: plan}, &update)

			updatePaths := test.paths
			if test.identity {
				updatePaths = append(updatePaths, "locale_id")
			}

			assert.Equal(t, updatePaths, attributeDiagnosticPaths(t, update.Diagnostics))
		})
	}
}

func TestLocaleModelDataPreservesKnownValuesAndExplicitNullFallback(t *testing.T) {
	t.Parallel()

	model := validLocaleRequestModel()
	model.Name = types.StringValue("")
	model.Code = types.StringValue("")
	model.ContentDeliveryAPI = types.BoolValue(false)
	model.ContentManagementAPI = types.BoolValue(false)
	model.Optional = types.BoolValue(false)

	request, diags := model.ToLocaleData()

	require.False(t, diags.HasError(), diags.Errors())
	assert.Empty(t, request.Name)
	assert.Empty(t, request.Code)
	assert.True(t, request.FallbackCode.IsNull())
	assert.False(t, request.ContentDeliveryApi)
	assert.False(t, request.ContentManagementApi)
	assert.False(t, request.Optional)

	model.FallbackCode = types.StringValue("")

	request, diags = model.ToLocaleData()

	require.False(t, diags.HasError(), diags.Errors())

	value, ok := request.FallbackCode.Get()
	require.True(t, ok)
	assert.Empty(t, value)
}
