package contentfulmanagement

type LocaleSys struct {
	Type    LocaleSysType `json:"type"`
	ID      string        `json:"id"`
	Version int           `json:"version"`
	Space   SpaceLink     `json:"space"`
}

type LocaleSysType string

const (
	LocaleSysTypeLocale LocaleSysType = "Locale"
)

func NewLocaleSys(spaceID, localeID string) LocaleSys {
	return LocaleSys{
		Type:    LocaleSysTypeLocale,
		ID:      localeID,
		Version: 1,
		Space:   NewSpaceLink(spaceID),
	}
}
