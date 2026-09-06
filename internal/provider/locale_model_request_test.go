package provider_test

import (
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

func TestLocaleModelParamsRejectUnknownAndNullRequiredStrings(t *testing.T) {
	t.Parallel()

	invalidValues := map[string]types.String{
		"unknown": types.StringUnknown(),
		"null":    types.StringNull(),
	}

	for attributeName, mutate := range map[string]func(*provider.LocaleModel, types.String){
		"space_id":       func(model *provider.LocaleModel, value types.String) { model.SpaceID = value },
		"environment_id": func(model *provider.LocaleModel, value types.String) { model.EnvironmentID = value },
	} {
		for valueName, value := range invalidValues {
			t.Run("create "+attributeName+" "+valueName, func(t *testing.T) {
				t.Parallel()

				model := validLocaleRequestModel()
				mutate(&model, value)

				params, diags := model.ToCreateLocaleParams(path.Empty())

				require.True(t, diags.HasError())
				assert.Empty(t, params)
				assert.Equal(t, []string{attributeName}, attributeDiagnosticPaths(t, diags))
			})
		}
	}

	for attributeName, mutate := range map[string]func(*provider.LocaleModel, types.String){
		"space_id":       func(model *provider.LocaleModel, value types.String) { model.SpaceID = value },
		"environment_id": func(model *provider.LocaleModel, value types.String) { model.EnvironmentID = value },
		"locale_id":      func(model *provider.LocaleModel, value types.String) { model.LocaleID = value },
	} {
		for valueName, value := range invalidValues {
			t.Run("update "+attributeName+" "+valueName, func(t *testing.T) {
				t.Parallel()

				model := validLocaleRequestModel()
				mutate(&model, value)

				params, diags := model.ToPutLocaleParams(path.Empty())

				require.True(t, diags.HasError())
				assert.Empty(t, params)
				assert.Equal(t, []string{attributeName}, attributeDiagnosticPaths(t, diags))
			})
		}
	}
}

func TestLocaleModelDataRejectsUnknownAndNullRequiredValues(t *testing.T) {
	t.Parallel()

	invalidStrings := map[string]types.String{
		"unknown": types.StringUnknown(),
		"null":    types.StringNull(),
	}

	for attributeName, mutate := range map[string]func(*provider.LocaleModel, types.String){
		"name": func(model *provider.LocaleModel, value types.String) { model.Name = value },
		"code": func(model *provider.LocaleModel, value types.String) { model.Code = value },
	} {
		for valueName, value := range invalidStrings {
			t.Run(attributeName+" "+valueName, func(t *testing.T) {
				t.Parallel()

				model := validLocaleRequestModel()
				mutate(&model, value)

				request, diags := model.ToLocaleData(path.Empty())

				require.True(t, diags.HasError())
				assert.Empty(t, request)
				assert.Equal(t, []string{attributeName}, attributeDiagnosticPaths(t, diags))
			})
		}
	}

	invalidBools := map[string]types.Bool{
		"unknown": types.BoolUnknown(),
		"null":    types.BoolNull(),
	}

	for attributeName, mutate := range map[string]func(*provider.LocaleModel, types.Bool){
		"content_delivery_api":   func(model *provider.LocaleModel, value types.Bool) { model.ContentDeliveryAPI = value },
		"content_management_api": func(model *provider.LocaleModel, value types.Bool) { model.ContentManagementAPI = value },
		"optional":               func(model *provider.LocaleModel, value types.Bool) { model.Optional = value },
	} {
		for valueName, value := range invalidBools {
			t.Run(attributeName+" "+valueName, func(t *testing.T) {
				t.Parallel()

				model := validLocaleRequestModel()
				mutate(&model, value)

				request, diags := model.ToLocaleData(path.Empty())

				require.True(t, diags.HasError())
				assert.Empty(t, request)
				assert.Equal(t, []string{attributeName}, attributeDiagnosticPaths(t, diags))
			})
		}
	}

	model := validLocaleRequestModel()
	model.FallbackCode = types.StringUnknown()

	request, diags := model.ToLocaleData(path.Empty())

	require.True(t, diags.HasError())
	assert.Empty(t, request)
	assert.Equal(t, []string{"fallback_code"}, attributeDiagnosticPaths(t, diags))
}

func TestLocaleModelRequestConvertersAggregateDiagnostics(t *testing.T) {
	t.Parallel()

	model := validLocaleRequestModel()
	model.SpaceID = types.StringUnknown()
	model.EnvironmentID = types.StringNull()
	model.LocaleID = types.StringUnknown()

	params, paramsDiags := model.ToPutLocaleParams(path.Empty())

	require.True(t, paramsDiags.HasError())
	assert.Empty(t, params)
	assert.Equal(t, []string{"space_id", "environment_id", "locale_id"}, attributeDiagnosticPaths(t, paramsDiags))

	model.Name = types.StringUnknown()
	model.Code = types.StringNull()
	model.FallbackCode = types.StringUnknown()
	model.ContentDeliveryAPI = types.BoolNull()
	model.ContentManagementAPI = types.BoolUnknown()
	model.Optional = types.BoolNull()

	request, requestDiags := model.ToLocaleData(path.Empty())

	require.True(t, requestDiags.HasError())
	assert.Empty(t, request)
	assert.Equal(t, []string{
		"name",
		"code",
		"fallback_code",
		"content_delivery_api",
		"content_management_api",
		"optional",
	}, attributeDiagnosticPaths(t, requestDiags))
}

func TestLocaleModelDataPreservesKnownValuesAndExplicitNullFallback(t *testing.T) {
	t.Parallel()

	model := validLocaleRequestModel()
	model.Name = types.StringValue("")
	model.Code = types.StringValue("")
	model.ContentDeliveryAPI = types.BoolValue(false)
	model.ContentManagementAPI = types.BoolValue(false)
	model.Optional = types.BoolValue(false)

	request, diags := model.ToLocaleData(path.Empty())

	require.False(t, diags.HasError(), diags.Errors())
	assert.Empty(t, request.Name)
	assert.Empty(t, request.Code)
	assert.True(t, request.FallbackCode.IsNull())
	assert.False(t, request.ContentDeliveryApi)
	assert.False(t, request.ContentManagementApi)
	assert.False(t, request.Optional)

	model.FallbackCode = types.StringValue("")

	request, diags = model.ToLocaleData(path.Empty())

	require.False(t, diags.HasError(), diags.Errors())

	value, ok := request.FallbackCode.Get()
	require.True(t, ok)
	assert.Empty(t, value)
}
