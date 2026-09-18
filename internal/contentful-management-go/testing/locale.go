package cmtesting

import cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"

func NewLocaleFromData(spaceID, environmentID, localeID string, data cm.LocaleData, defaultLocale bool) cm.Locale {
	locale := cm.Locale{
		Sys: cm.LocaleSys{Space: cm.NewSpaceLink(spaceID), Environment: cm.NewEnvironmentLink(environmentID), Type: cm.LocaleSysTypeLocale, ID: localeID},
	}

	updateLocaleFromData(&locale, data)
	locale.Default = defaultLocale

	return locale
}

func updateLocaleFromData(locale *cm.Locale, data cm.LocaleData) {
	locale.Sys.Version.SetTo(locale.Sys.Version.Or(0) + 1)

	locale.Name = data.Name
	locale.Code = data.Code

	locale.FallbackCode = cm.NewOptNilPointerString(data.FallbackCode.ValueStringPointer())

	locale.ContentDeliveryApi = data.ContentDeliveryApi
	locale.ContentManagementApi = data.ContentManagementApi
	locale.Optional = data.Optional
}

func localeDataEqual(locale *cm.Locale, data *cm.LocaleData) bool {
	return locale.Name == data.Name && locale.Code == data.Code &&
		locale.FallbackCode == cm.NewOptNilPointerString(data.FallbackCode.ValueStringPointer()) &&
		locale.ContentDeliveryApi == data.ContentDeliveryApi &&
		locale.ContentManagementApi == data.ContentManagementApi && locale.Optional == data.Optional
}

func projectLocaleResponse(locale cm.Locale, environmentID string) cm.Locale {
	locale.Sys.Environment = cm.NewEnvironmentLink(environmentID)

	return locale
}
