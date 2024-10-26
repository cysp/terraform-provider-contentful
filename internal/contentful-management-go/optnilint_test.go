package client_test

import (
	"testing"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
)

func TestNewOptNilPointerInt(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    *int
		expected contentfulManagement.OptNilInt
	}{
		"test": {
			input:    addressOf(42),
			expected: contentfulManagement.NewOptNilInt(42),
		},
		"zero": {
			input:    addressOf(0),
			expected: contentfulManagement.NewOptNilInt(0),
		},
		"nil": {
			input:    nil,
			expected: contentfulManagement.NewOptNilIntNull(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := contentfulManagement.NewOptNilPointerInt(test.input)

			assert.EqualValues(t, test.expected, actual)
		})
	}
}

func TestNewOptNilPointerInt64(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    *int64
		expected contentfulManagement.OptNilInt
	}{
		"test": {
			input:    addressOf(int64(42)),
			expected: contentfulManagement.NewOptNilInt(42),
		},
		"zero": {
			input:    addressOf(int64(0)),
			expected: contentfulManagement.NewOptNilInt(0),
		},
		"nil": {
			input:    nil,
			expected: contentfulManagement.NewOptNilIntNull(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := contentfulManagement.NewOptNilPointerInt64(test.input)

			assert.EqualValues(t, test.expected, actual)
		})
	}
}

func TestOptNilIntValueIntPointer(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    contentfulManagement.OptNilInt
		expected *int
	}{
		"test": {
			input:    contentfulManagement.NewOptNilInt(42),
			expected: addressOf(42),
		},
		"zero": {
			input:    contentfulManagement.NewOptNilInt(0),
			expected: addressOf(0),
		},
		"null": {
			input:    contentfulManagement.NewOptNilIntNull(),
			expected: nil,
		},
		"nil": {
			input:    contentfulManagement.OptNilInt{},
			expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := test.input.ValueIntPointer()

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestOptNilIntValueInt64Pointer(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    contentfulManagement.OptNilInt
		expected *int64
	}{
		"test": {
			input:    contentfulManagement.NewOptNilInt(42),
			expected: addressOf(int64(42)),
		},
		"zero": {
			input:    contentfulManagement.NewOptNilInt(0),
			expected: addressOf(int64(0)),
		},
		"null": {
			input:    contentfulManagement.NewOptNilIntNull(),
			expected: nil,
		},
		"nil": {
			input:    contentfulManagement.OptNilInt{},
			expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := test.input.ValueInt64Pointer()

			assert.Equal(t, test.expected, actual)
		})
	}
}
