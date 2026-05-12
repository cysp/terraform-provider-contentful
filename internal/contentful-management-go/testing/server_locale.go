package cmtesting

import cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"

func (s *Server) SetLocale(spaceID, environmentID, localeID string, data cm.LocaleData, defaultLocale bool) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	s.h.registerSpaceEnvironment(spaceID, environmentID, "ready")

	locale := NewLocaleFromData(spaceID, environmentID, localeID, data, defaultLocale)
	s.h.locales.Set(spaceID, environmentID, localeID, &locale)
}
