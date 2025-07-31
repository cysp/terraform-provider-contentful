package contentfulmanagement_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
)

func TestNewOptNilPointerString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    *string
		expected cm.OptNilString
	}{
		"test": {
			input:    addressOf("test"),
			expected: cm.NewOptNilString("test"),
		},
		"empty": {
			input:    addressOf(""),
			expected: cm.NewOptNilString(""),
		},
		"nil": {
			input:    nil,
			expected: cm.NewOptNilStringNull(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := cm.NewOptNilPointerString(test.input)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestOptNilStringValueStringPointer(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    cm.OptNilString
		expected *string
	}{
		"test": {
			input:    cm.NewOptNilString("test"),
			expected: addressOf("test"),
		},
		"empty": {
			input:    cm.NewOptNilString(""),
			expected: addressOf(""),
		},
		"null": {
			input:    cm.NewOptNilStringNull(),
			expected: nil,
		},
		"nil": {
			input:    cm.OptNilString{},
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
