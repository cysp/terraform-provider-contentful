package contentfulmanagement_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
)

func TestNewOptNilPointerInt64(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    *int64
		expected cm.OptNilInt
	}{
		"test": {
			input:    new(int64(42)),
			expected: cm.NewOptNilInt(42),
		},
		"zero": {
			input:    new(int64(0)),
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

			assert.Equal(t, test.expected, actual)
		})
	}
}
