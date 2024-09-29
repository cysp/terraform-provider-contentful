package client_test

import (
	"testing"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
)

func TestOptNilStringValueStringPointer(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    contentfulManagement.OptNilString
		expected *string
	}{
		"test": {
			input:    contentfulManagement.NewOptNilString("test"),
			expected: addressOf("test"),
		},
		"empty": {
			input:    contentfulManagement.NewOptNilString(""),
			expected: addressOf(""),
		},
		"null": {
			input:    contentfulManagement.NewOptNilStringNull(),
			expected: nil,
		},
		"nil": {
			input:    contentfulManagement.OptNilString{},
			expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := test.input.ValueStringPointer()

			assert.Equal(t, test.expected, actual)
		})
	}
}
