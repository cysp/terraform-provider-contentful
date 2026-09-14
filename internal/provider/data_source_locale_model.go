package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type LocaleDataSourceItemModel struct {
	LocaleID             types.String `tfsdk:"locale_id"`
	Name                 types.String `tfsdk:"name"`
	Code                 types.String `tfsdk:"code"`
	Default              types.Bool   `tfsdk:"default"`
	FallbackCode         types.String `tfsdk:"fallback_code"`
	Optional             types.Bool   `tfsdk:"optional"`
	ContentManagementAPI types.Bool   `tfsdk:"content_management_api"`
	ContentDeliveryAPI   types.Bool   `tfsdk:"content_delivery_api"`
}
type LocaleDataSourceModel struct {
	IDIdentityModel
	LocaleDataSourceItemModel

	SpaceID       types.String   `tfsdk:"space_id"`
	EnvironmentID types.String   `tfsdk:"environment_id"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}
type LocalesDataSourceModel struct {
	IDIdentityModel

	SpaceID       types.String                `tfsdk:"space_id"`
	EnvironmentID types.String                `tfsdk:"environment_id"`
	Locales       []LocaleDataSourceItemModel `tfsdk:"locales"`
	Timeouts      timeouts.Value              `tfsdk:"timeouts"`
}
