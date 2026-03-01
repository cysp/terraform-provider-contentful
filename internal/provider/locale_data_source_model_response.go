package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewLocaleDataSourceModelFromResponse(_ context.Context, locale cm.Locale) (LocaleDataSourceModel, diag.Diagnostics) {
	spaceID := locale.Sys.Space.Sys.ID
	localeID := locale.Sys.ID

	return LocaleDataSourceModel{
		ID:                   NewIDIdentityModelFromMultipartID(spaceID, localeID).ID,
		SpaceID:              types.StringValue(spaceID),
		LocaleID:             types.StringValue(localeID),
		Name:                 types.StringValue(locale.Name),
		Code:                 types.StringValue(locale.Code),
		FallbackCode:         util.OptNilStringToStringValue(locale.FallbackCode),
		Optional:             types.BoolValue(locale.Optional),
		Default:              types.BoolValue(locale.Default),
		ContentDeliveryAPI:   types.BoolValue(locale.ContentDeliveryAPI),
		ContentManagementAPI: types.BoolValue(locale.ContentManagementAPI),
	}, diag.Diagnostics{}
}
