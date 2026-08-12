package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/stretchr/testify/require"
)

func TestNormalizeJSON(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    string
		expected string
	}{
		"already normalized": {
			input:    `{"key":"value"}`,
			expected: `{"key":"value"}`,
		},
		"extra whitespace": {
			input:    `{  "key"  :  "value"  }`,
			expected: `{"key":"value"}`,
		},
		"different property order": {
			input:    `{"b":"2","a":"1"}`,
			expected: `{"a":"1","b":"2"}`,
		},
		"nested object": {
			input:    `{"outer":{"inner":"value"}}`,
			expected: `{"outer":{"inner":"value"}}`,
		},
		"array": {
			input:    `[1,2,3]`,
			expected: `[1,2,3]`,
		},
		"integer beyond float64 exact range": {
			input:    `{"value":9007199254740993}`,
			expected: `{"value":9007199254740993}`,
		},
		"high precision decimal": {
			input:    `{"value":0.12345678901234567890123456789}`,
			expected: `{"value":0.12345678901234567890123456789}`,
		},
		"exponent notation": {
			input:    `{"value":1.234567890123456789e+100}`,
			expected: `{"value":1.234567890123456789e+100}`,
		},
		"invalid JSON": {
			input:    `{invalid}`,
			expected: `{invalid}`,
		},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := NormalizeJSON([]byte(testCase.input))

			require.Equal(t, testCase.expected, actual)
		})
	}
}
