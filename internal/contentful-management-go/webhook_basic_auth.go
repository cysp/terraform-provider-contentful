package contentfulmanagement

import "unicode/utf16"

const (
	webhookBasicUsernameLimit = 50
	webhookBasicPasswordLimit = 255
)

// WebhookBasicUsernameTooLong reports whether the username exceeds the CMA limit in UTF-16 code units.
func WebhookBasicUsernameTooLong(username string) bool {
	return exceedsUTF16Length(username, webhookBasicUsernameLimit)
}

func WebhookBasicPasswordTooLong(password string) bool {
	return exceedsUTF16Length(password, webhookBasicPasswordLimit)
}

func exceedsUTF16Length(value string, limit int) bool {
	for _, r := range value {
		limit -= utf16.RuneLen(r)
		if limit < 0 {
			return true
		}
	}

	return false
}
