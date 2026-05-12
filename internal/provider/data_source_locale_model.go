package provider

import (
	datasourcetimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type LocaleDataSourceModel struct {
	IDIdentityModel
	LocaleIdentityModel

	Name                 types.String `tfsdk:"name"`
	Code                 types.String `tfsdk:"code"`
	FallbackCode         types.String `tfsdk:"fallback_code"`
	ContentDeliveryAPI   types.Bool   `tfsdk:"content_delivery_api"`
	ContentManagementAPI types.Bool   `tfsdk:"content_management_api"`
	Optional             types.Bool   `tfsdk:"optional"`
	Default              types.Bool   `tfsdk:"default"`
	InternalCode         types.String `tfsdk:"internal_code"`

	Timeouts datasourcetimeouts.Value `tfsdk:"timeouts"`
}
