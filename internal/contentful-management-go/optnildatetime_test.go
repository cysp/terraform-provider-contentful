package client_test

import (
	"testing"
	"time"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
)

func TestOptNilDateTimeValueTimePointer(t *testing.T) {
	t.Parallel()

	epoch := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)

	tests := map[string]struct {
		input    contentfulManagement.OptNilDateTime
		expected *time.Time
	}{
		"valid": {
			input:    contentfulManagement.NewOptNilDateTime(epoch),
			expected: addressOf(epoch),
		},
		"null": {
			input:    contentfulManagement.NewOptNilDateTimeNull(),
			expected: nil,
		},
		"nil": {
			input:    contentfulManagement.OptNilDateTime{},
			expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := test.input.ValueTimePointer()

			assert.Equal(t, test.expected, actual)
		})
	}
}
