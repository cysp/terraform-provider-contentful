package client_test

import (
	"testing"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
)

func TestNewOptPointerString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    *string
		expected contentfulManagement.OptString
	}{
		"test": {
			input:    addressOf("test"),
			expected: contentfulManagement.NewOptString("test"),
		},
		"empty": {
			input:    addressOf(""),
			expected: contentfulManagement.NewOptString(""),
		},
		"nil": {
			input:    nil,
			expected: contentfulManagement.OptString{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := contentfulManagement.NewOptPointerString(test.input)

			assert.EqualValues(t, test.expected, actual)
		})
	}
}

func TestOptStringValueStringPointer(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    contentfulManagement.OptString
		expected *string
	}{
		"test": {
			input:    contentfulManagement.NewOptString("test"),
			expected: addressOf("test"),
		},
		"empty": {
			input:    contentfulManagement.NewOptString(""),
			expected: addressOf(""),
		},
		"nil": {
			input:    contentfulManagement.OptString{},
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
