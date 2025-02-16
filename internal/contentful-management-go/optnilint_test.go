package client_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
)

func TestNewOptNilPointerInt(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    *int
		expected cm.OptNilInt
	}{
		"test": {
			input:    addressOf(42),
			expected: cm.NewOptNilInt(42),
		},
		"zero": {
			input:    addressOf(0),
			expected: cm.NewOptNilInt(0),
		},
		"nil": {
			input:    nil,
			expected: cm.NewOptNilIntNull(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := cm.NewOptNilPointerInt(test.input)

			assert.EqualValues(t, test.expected, actual)
		})
	}
}

func TestNewOptNilPointerInt64(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    *int64
		expected cm.OptNilInt
	}{
		"test": {
			input:    addressOf(int64(42)),
			expected: cm.NewOptNilInt(42),
		},
		"zero": {
			input:    addressOf(int64(0)),
			expected: cm.NewOptNilInt(0),
		},
		"nil": {
			input:    nil,
			expected: cm.NewOptNilIntNull(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := cm.NewOptNilPointerInt64(test.input)

			assert.EqualValues(t, test.expected, actual)
		})
	}
}

func TestOptNilIntValueIntPointer(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    cm.OptNilInt
		expected *int
	}{
		"test": {
			input:    cm.NewOptNilInt(42),
			expected: addressOf(42),
		},
		"zero": {
			input:    cm.NewOptNilInt(0),
			expected: addressOf(0),
		},
		"null": {
			input:    cm.NewOptNilIntNull(),
			expected: nil,
		},
		"nil": {
			input:    cm.OptNilInt{},
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
		input    cm.OptNilInt
		expected *int64
	}{
		"test": {
			input:    cm.NewOptNilInt(42),
			expected: addressOf(int64(42)),
		},
		"zero": {
			input:    cm.NewOptNilInt(0),
			expected: addressOf(int64(0)),
		},
		"null": {
			input:    cm.NewOptNilIntNull(),
			expected: nil,
		},
		"nil": {
			input:    cm.OptNilInt{},
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
