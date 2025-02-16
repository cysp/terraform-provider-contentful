package client_test

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
			input:    addressOf(true),
			expected: cm.NewOptBool(true),
		},
		"false": {
			input:    addressOf(false),
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

			assert.EqualValues(t, test.expected, actual)
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
			expected: addressOf(true),
		},
		"false": {
			input:    cm.NewOptBool(false),
			expected: addressOf(false),
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
