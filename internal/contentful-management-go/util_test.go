package client_test

import (
	"testing"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
)

func addressOf[T any](value T) *T {
	return &value
}

func TestNewOptPointerBool(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    *bool
		expected contentfulManagement.OptBool
	}{
		"true": {
			input:    addressOf(true),
			expected: contentfulManagement.NewOptBool(true),
		},
		"false": {
			input:    addressOf(false),
			expected: contentfulManagement.NewOptBool(false),
		},
		"nil": {
			input:    nil,
			expected: contentfulManagement.OptBool{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := contentfulManagement.NewOptPointerBool(test.input)

			assert.EqualValues(t, test.expected, actual)
		})
	}
}

func TestOptBoolValueBoolPointer(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    contentfulManagement.OptBool
		expected *bool
	}{
		"true": {
			input:    contentfulManagement.NewOptBool(true),
			expected: addressOf(true),
		},
		"false": {
			input:    contentfulManagement.NewOptBool(false),
			expected: addressOf(false),
		},
		"nil": {
			input:    contentfulManagement.OptBool{},
			expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := test.input.ValueBoolPointer()

			assert.Equal(t, test.expected, actual)
		})
	}
}

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
