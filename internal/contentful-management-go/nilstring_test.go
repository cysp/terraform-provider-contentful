package contentfulmanagement_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
)

func TestNewNilPointerString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    *string
		expected cm.NilString
	}{
		"test": {
			input:    addressOf("test"),
			expected: cm.NewNilString("test"),
		},
		"empty": {
			input:    addressOf(""),
			expected: cm.NewNilString(""),
		},
		"nil": {
			input:    nil,
			expected: cm.NewNilStringNull(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := cm.NewNilPointerString(test.input)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestNilStringValueStringPointer(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    cm.NilString
		expected *string
	}{
		"test": {
			input:    cm.NewNilString("test"),
			expected: addressOf("test"),
		},
		"empty": {
			input:    cm.NewNilString(""),
			expected: addressOf(""),
		},
		"null": {
			input:    cm.NewNilStringNull(),
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
