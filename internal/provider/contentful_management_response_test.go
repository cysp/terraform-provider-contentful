//nolint:testpackage // The package-private CMA response classifier is intentionally tested through its interface.
package provider

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStatusCodeResponse int

func (response testStatusCodeResponse) GetStatusCode() int {
	return int(response)
}

func TestContentfulResponseIsNotFound(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		response any
		expected bool
	}{
		"not found": {
			response: testStatusCodeResponse(http.StatusNotFound),
			expected: true,
		},
		"other status": {
			response: testStatusCodeResponse(http.StatusConflict),
		},
		"non-status response": {
			response: struct{}{},
		},
		"nil response": {},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, contentfulResponseIsNotFound(test.response))
		})
	}
}
