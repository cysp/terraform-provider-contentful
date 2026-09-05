package contentfulmanagement_test

import (
	"encoding/json"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocaleDecodesDocumentedContentfulResponse(t *testing.T) {
	t.Parallel()

	response := []byte(`{
		"name": "English (United States)",
		"code": "en-US",
		"fallbackCode": "en-US",
		"contentDeliveryApi": true,
		"contentManagementApi": true,
		"default": false,
		"optional": false,
		"sys": {
			"type": "Locale",
			"id": "34N35DoyUQAtaKwWTgZs34",
			"version": 1,
			"space": {
				"sys": {
					"type": "Link",
					"linkType": "Space",
					"id": "yadj1kx9rmg0"
				}
			},
			"environment": {
				"sys": {
					"type": "Link",
					"linkType": "Environment",
					"id": "staging"
				}
			}
		}
	}`)

	var locale cm.Locale
	require.NoError(t, json.Unmarshal(response, &locale))

	assert.Equal(t, "34N35DoyUQAtaKwWTgZs34", locale.Sys.ID)
	assert.Equal(t, 1, locale.Sys.Version)
	assert.Equal(t, "en-US", locale.Code)
}
