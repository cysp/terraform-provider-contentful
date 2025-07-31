package contentfulmanagement_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
)

func TestNewOptPointerString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    *string
		expected cm.OptString
	}{
		"test": {
			input:    addressOf("test"),
			expected: cm.NewOptString("test"),
		},
		"empty": {
			input:    addressOf(""),
			expected: cm.NewOptString(""),
		},
		"nil": {
			input:    nil,
			expected: cm.OptString{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := cm.NewOptPointerString(test.input)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestOptStringValueStringPointer(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    cm.OptString
		expected *string
	}{
		"test": {
			input:    cm.NewOptString("test"),
			expected: addressOf("test"),
		},
		"empty": {
			input:    cm.NewOptString(""),
			expected: addressOf(""),
		},
		"nil": {
			input:    cm.OptString{},
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
