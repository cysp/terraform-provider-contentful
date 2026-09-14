package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocaleMutationResponseChecksAllPlannedValues(t *testing.T) {
	t.Parallel()

	for field, mutate := range map[string]func(*cm.Locale){
		"space_id":               func(v *cm.Locale) { v.Sys.Space = cm.NewSpaceLink("other") },
		"environment_id":         func(v *cm.Locale) { v.Sys.Environment = cm.NewEnvironmentLink("other") },
		"locale_id":              func(v *cm.Locale) { v.Sys.ID = "other" },
		"name":                   func(v *cm.Locale) { v.Name = "Other" },
		"code":                   func(v *cm.Locale) { v.Code = "other" },
		"fallback_code":          func(v *cm.Locale) { v.FallbackCode = cm.NewOptNilString("") },
		"content_delivery_api":   func(v *cm.Locale) { v.ContentDeliveryApi = false },
		"content_management_api": func(v *cm.Locale) { v.ContentManagementApi = false },
		"optional":               func(v *cm.Locale) { v.Optional = true },
		"default":                func(v *cm.Locale) { v.Default = true },
	} {
		t.Run(field, func(t *testing.T) {
			t.Parallel()

			plan := validLocaleRequestModel()
			plan.ID = types.StringValue("space/environment/locale")
			plan.Default = types.BoolValue(false)
			response := cmt.NewLocaleFromData("space", "environment", "locale", cm.LocaleData{Name: "English (Australia)", Code: "en-AU", FallbackCode: cm.NewNilStringNull(), ContentDeliveryApi: true, ContentManagementApi: true, Optional: false}, false)
			_, diagnostics := provider.ReconcileLocaleMutationResponse(response, plan)
			require.Empty(t, diagnostics)
			mutate(&response)
			state, diagnostics := provider.ReconcileLocaleMutationResponse(response, plan)
			require.Len(t, diagnostics, 1)
			assert.Equal(t, []string{field}, attributeDiagnosticPaths(t, diagnostics))
			assert.Equal(t, plan.LocaleIdentityModel, state.LocaleIdentityModel)
			assert.Equal(t, plan.ID, state.ID)
		})
	}
}
