package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocaleMutationResponseChecksAllPlannedValues(t *testing.T) {
	t.Parallel()

	plan := validLocaleRequestModel()
	plan.ID = types.StringValue("space/environment/locale")
	plan.Default = types.BoolValue(false)
	response := cm.Locale{
		Sys: cm.LocaleSys{
			ID: "locale", Space: cm.NewSpaceLink("space"), Environment: cm.NewEnvironmentLink("environment"),
		},
		Name:                 "English (Australia)",
		Code:                 "en-AU",
		FallbackCode:         cm.NewOptNilStringNull(),
		ContentDeliveryApi:   true,
		ContentManagementApi: true,
	}
	_, diagnostics := provider.ReconcileLocaleMutationResponse(response, plan)
	require.Empty(t, diagnostics)

	for field, test := range map[string]struct {
		mutate func(*cm.Locale)
		want   any
	}{
		"space_id":               {func(v *cm.Locale) { v.Sys.Space = cm.NewSpaceLink("other") }, "space"},
		"environment_id":         {func(v *cm.Locale) { v.Sys.Environment = cm.NewEnvironmentLink("other") }, "environment"},
		"locale_id":              {func(v *cm.Locale) { v.Sys.ID = "other" }, "locale"},
		"name":                   {func(v *cm.Locale) { v.Name = "Other" }, "Other"},
		"code":                   {func(v *cm.Locale) { v.Code = "other" }, "other"},
		"fallback_code":          {func(v *cm.Locale) { v.FallbackCode = cm.NewOptNilString("") }, ""},
		"content_delivery_api":   {func(v *cm.Locale) { v.ContentDeliveryApi = false }, false},
		"content_management_api": {func(v *cm.Locale) { v.ContentManagementApi = false }, false},
		"optional":               {func(v *cm.Locale) { v.Optional = true }, true},
		"default":                {func(v *cm.Locale) { v.Default = true }, true},
	} {
		t.Run(field, func(t *testing.T) {
			t.Parallel()

			contradiction := response
			test.mutate(&contradiction)
			state, diagnostics := provider.ReconcileLocaleMutationResponse(contradiction, plan)
			require.Len(t, diagnostics, 1)
			assert.Equal(t, []string{field}, attributeDiagnosticPaths(t, diagnostics))
			assert.Equal(t, plan.LocaleIdentityModel, state.LocaleIdentityModel)
			assert.Equal(t, plan.ID, state.ID)

			result := tfsdk.State{Schema: provider.LocaleResourceSchema(t.Context())}
			require.Empty(t, result.Set(t.Context(), state))

			var attributes map[string]tftypes.Value
			require.NoError(t, result.Raw.As(&attributes))
			assert.Equal(t, tftypes.NewValue(attributes[field].Type(), test.want), attributes[field])
		})
	}
}
