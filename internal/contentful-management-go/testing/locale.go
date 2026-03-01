package cmtesting

import cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"

func NewLocaleFromRequest(spaceID, localeID string, req *cm.LocaleRequest) cm.Locale {
	locale := cm.Locale{
		Sys:                  cm.NewLocaleSys(spaceID, localeID),
		Name:                 req.Name,
		Code:                 req.Code,
		Optional:             false,
		ContentDeliveryAPI:   true,
		ContentManagementAPI: true,
	}

	UpdateLocaleFromRequest(&locale, req)

	return locale
}

func UpdateLocaleFromRequest(locale *cm.Locale, req *cm.LocaleRequest) {
	locale.Sys.Version++

	locale.Name = req.Name

	if req.Code != "" {
		locale.Code = req.Code
	}

	if req.FallbackCode.IsSet() {
		locale.FallbackCode = req.FallbackCode
	}

	if value, ok := req.Optional.Get(); ok {
		locale.Optional = value
	}

	if value, ok := req.Default.Get(); ok {
		locale.Default = value
	}

	if value, ok := req.ContentDeliveryAPI.Get(); ok {
		locale.ContentDeliveryAPI = value
	}

	if value, ok := req.ContentManagementAPI.Get(); ok {
		locale.ContentManagementAPI = value
	}
}
