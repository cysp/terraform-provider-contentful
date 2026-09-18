package cmtesting

import cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"

func projectLocaleResponse(locale cm.Locale, environmentID string) cm.Locale {
	locale.Sys.Environment = cm.NewEnvironmentLink(environmentID)

	return locale
}
