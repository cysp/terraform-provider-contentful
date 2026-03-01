package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type LocaleIdentityModel struct {
	SpaceID  types.String `tfsdk:"space_id"`
	LocaleID types.String `tfsdk:"locale_id"`
}

type LocaleModel struct {
	IDIdentityModel
	LocaleIdentityModel

	Name                 types.String `tfsdk:"name"`
	Code                 types.String `tfsdk:"code"`
	FallbackCode         types.String `tfsdk:"fallback_code"`
	Optional             types.Bool   `tfsdk:"optional"`
	Default              types.Bool   `tfsdk:"default"`
	ContentDeliveryAPI   types.Bool   `tfsdk:"content_delivery_api"`
	ContentManagementAPI types.Bool   `tfsdk:"content_management_api"`
}
