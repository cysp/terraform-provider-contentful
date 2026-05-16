package cmtesting

import cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"

func NewLocaleFromData(spaceID, environmentID, localeID string, data cm.LocaleData, defaultLocale bool) cm.Locale {
	locale := cm.Locale{
		Sys: cm.NewLocaleSys(spaceID, environmentID, localeID),
	}

	UpdateLocaleFromData(&locale, data)
	locale.Default = defaultLocale
	locale.InternalCode = data.Code

	return locale
}

func UpdateLocaleFromData(locale *cm.Locale, data cm.LocaleData) {
	locale.Sys.Version++

	locale.Name = data.Name
	locale.Code = data.Code
	locale.FallbackCode = data.FallbackCode
	locale.ContentDeliveryApi = data.ContentDeliveryApi
	locale.ContentManagementApi = data.ContentManagementApi
	locale.Optional = data.Optional
	locale.InternalCode = data.Code
}
