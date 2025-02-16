package client_test

import (
	"testing"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
)

func TestOptNilDateTimeValueTimePointer(t *testing.T) {
	t.Parallel()

	epoch := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)

	tests := map[string]struct {
		input    cm.OptNilDateTime
		expected *time.Time
	}{
		"valid": {
			input:    cm.NewOptNilDateTime(epoch),
			expected: addressOf(epoch),
		},
		"null": {
			input:    cm.NewOptNilDateTimeNull(),
			expected: nil,
		},
		"nil": {
			input:    cm.OptNilDateTime{},
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
