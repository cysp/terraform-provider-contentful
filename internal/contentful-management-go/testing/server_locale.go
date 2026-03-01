package cmtesting

import cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"

func (s *Server) SetLocale(spaceID, localeID string, request cm.LocaleRequest) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	locale := NewLocaleFromRequest(spaceID, localeID, &request)
	s.h.locales.Set(spaceID, localeID, &locale)
}
