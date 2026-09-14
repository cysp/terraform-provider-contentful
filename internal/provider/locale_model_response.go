package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewLocaleResourceModelFromResponse(locale cm.Locale) LocaleModel {
	spaceID := locale.Sys.Space.Sys.ID
	environmentID := locale.Sys.Environment.Sys.ID
	localeID := locale.Sys.ID

	return LocaleModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, environmentID, localeID),
		LocaleIdentityModel: LocaleIdentityModel{
			SpaceID:       types.StringValue(spaceID),
			EnvironmentID: types.StringValue(environmentID),
			LocaleID:      types.StringValue(localeID),
		},
		Name:                 types.StringValue(locale.Name),
		Code:                 types.StringValue(locale.Code),
		FallbackCode:         types.StringPointerValue(locale.FallbackCode.ValueStringPointer()),
		ContentDeliveryAPI:   types.BoolValue(locale.ContentDeliveryApi),
		ContentManagementAPI: types.BoolValue(locale.ContentManagementApi),
		Optional:             types.BoolValue(locale.Optional),
		Default:              types.BoolValue(locale.Default),
		Timeouts:             TimeoutsNull(),
	}
}

// ReconcileLocaleMutationResponse retains endpoint identity and reports any
// contradiction to a known planned value alongside the returned recovery state.
func ReconcileLocaleMutationResponse(locale cm.Locale, plan LocaleModel) (LocaleModel, diag.Diagnostics) {
	data := NewLocaleResourceModelFromResponse(locale)
	reconciler := mutationResponseReconciler{resourceName: "locale"}
	reconciler.pinIdentity(path.Root("space_id"), plan.SpaceID, data.SpaceID, &data.SpaceID)
	reconciler.pinIdentity(path.Root("environment_id"), plan.EnvironmentID, data.EnvironmentID, &data.EnvironmentID)
	reconciler.pinIdentity(path.Root("locale_id"), plan.LocaleID, data.LocaleID, &data.LocaleID)
	data.IDIdentityModel = NewIDIdentityModelFromMultipartID(data.SpaceID.ValueString(), data.EnvironmentID.ValueString(), data.LocaleID.ValueString())
	reconciler.compareExact(path.Root("id"), "Contentful returned a different locale id", plan.ID, data.ID)
	reconciler.compareExact(path.Root("name"), "Contentful returned a different locale name", plan.Name, data.Name)
	reconciler.compareExact(path.Root("code"), "Contentful returned a different locale code", plan.Code, data.Code)
	reconciler.compareExact(path.Root("fallback_code"), "Contentful returned a different locale fallback code", plan.FallbackCode, data.FallbackCode)
	reconciler.compareExact(path.Root("content_delivery_api"), "Contentful returned a different locale content delivery api", plan.ContentDeliveryAPI, data.ContentDeliveryAPI)
	reconciler.compareExact(path.Root("content_management_api"), "Contentful returned a different locale content management api", plan.ContentManagementAPI, data.ContentManagementAPI)
	reconciler.compareExact(path.Root("optional"), "Contentful returned a different locale optional", plan.Optional, data.Optional)
	reconciler.compareExact(path.Root("default"), "Contentful returned a different locale default", plan.Default, data.Default)

	return data, reconciler.diagnostics
}
