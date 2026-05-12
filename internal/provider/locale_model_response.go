package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewLocaleResourceModelFromResponse(_ context.Context, locale cm.Locale) (LocaleModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	spaceID := locale.Sys.Space.Sys.ID
	environmentID := locale.Sys.Environment.Sys.ID
	localeID := locale.Sys.ID

	fallbackCode := types.StringNull()
	if value, ok := locale.FallbackCode.Get(); ok {
		fallbackCode = types.StringValue(value)
	}

	model := LocaleModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, environmentID, localeID),
		LocaleIdentityModel: LocaleIdentityModel{
			SpaceID:       types.StringValue(spaceID),
			EnvironmentID: types.StringValue(environmentID),
			LocaleID:      types.StringValue(localeID),
		},
		Name:                 types.StringValue(locale.Name),
		Code:                 types.StringValue(locale.Code),
		FallbackCode:         fallbackCode,
		ContentDeliveryAPI:   types.BoolValue(locale.ContentDeliveryApi),
		ContentManagementAPI: types.BoolValue(locale.ContentManagementApi),
		Optional:             types.BoolValue(locale.Optional),
		Default:              types.BoolValue(locale.Default),
		InternalCode:         types.StringValue(locale.InternalCode),
	}

	return model, diags
}
