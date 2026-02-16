package contentfulmanagement_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
)

func TestNewOptPointerBool(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    *bool
		expected cm.OptBool
	}{
		"true": {
			input:    new(true),
			expected: cm.NewOptBool(true),
		},
		"false": {
			input:    new(false),
			expected: cm.NewOptBool(false),
		},
		"nil": {
			input:    nil,
			expected: cm.OptBool{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := cm.NewOptPointerBool(test.input)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestOptBoolValueBoolPointer(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    cm.OptBool
		expected *bool
	}{
		"true": {
			input:    cm.NewOptBool(true),
			expected: new(true),
		},
		"false": {
			input:    cm.NewOptBool(false),
			expected: new(false),
		},
		"nil": {
			input:    cm.OptBool{},
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
