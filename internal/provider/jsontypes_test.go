package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
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
		"empty array": {
			input:    `[ ]`,
			expected: `[]`,
		},
		"empty object": {
			input:    `{ }`,
			expected: `{}`,
		},
		"string": {
			input:    ` "value" `,
			expected: `"value"`,
		},
		"boolean": {
			input:    ` true `,
			expected: `true`,
		},
		"null": {
			input:    ` null `,
			expected: `null`,
		},
		"keys and values escape HTML and control characters": {
			input:    `{"a<b&c>d\"\\\n":"e<f&g>h\"\\\t"}`,
			expected: `{"a\u003cb\u0026c\u003ed\"\\\n":"e\u003cf\u0026g\u003eh\"\\\t"}`,
		},
		"nested objects and arrays": {
			input:    `{"z":[{"z":"<last>","a&b":"first&"},"<second>"],"a":{"<key>":">value"}}`,
			expected: `{"a":{"\u003ckey\u003e":"\u003evalue"},"z":[{"a\u0026b":"first\u0026","z":"\u003clast\u003e"},"\u003csecond\u003e"]}`,
		},
		"duplicate object keys retain last value": {
			input:    `{"key":"first","key":"last"}`,
			expected: `{"key":"last"}`,
		},
		"invalid UTF-8 string uses replacement character": {
			input:    "\"\xff\"",
			expected: `"�"`,
		},
		"invalid UTF-8 key collision retains last value": {
			input:    "{\"\xff\":\"invalid\",\"�\":\"replacement\"}",
			expected: `{"�":"replacement"}`,
		},
		"unpaired surrogate uses replacement character": {
			input:    `{"value":"\ud800"}`,
			expected: `{"value":"�"}`,
		},
		"unpaired surrogate key collision retains last value": {
			input:    `{"\ud800":"surrogate","�":"replacement"}`,
			expected: `{"�":"replacement"}`,
		},
		"paired surrogate": {
			input:    `"\ud83d\ude00"`,
			expected: `"😀"`,
		},
		"line and paragraph separators": {
			input:    "\"\u2028\u2029\"",
			expected: `"\u2028\u2029"`,
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
		"negative zero": {
			input:    `{"value":-0}`,
			expected: `{"value":-0}`,
		},
		"invalid JSON": {
			input:    `{invalid}`,
			expected: `{invalid}`,
		},
		"trailing JSON": {
			input:    `{"value":1} {"value":2}`,
			expected: `{"value":1} {"value":2}`,
		},
		"trailing malformed JSON": {
			input:    `true {`,
			expected: `true {`,
		},
		"invalid nested object": {
			input:    `{"a":{"b":invalid}}`,
			expected: `{"a":{"b":invalid}}`,
		},
		"invalid nested array": {
			input:    `{"a":[true,invalid]}`,
			expected: `{"a":[true,invalid]}`,
		},
		"incomplete nested structure": {
			input:    `{"a":[{"b":true}`,
			expected: `{"a":[{"b":true}`,
		},
		"empty input": {},
		"whitespace-only input": {
			input:    " \n\t ",
			expected: " \n\t ",
		},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := NormalizeJSON([]byte(testCase.input))

			require.Equal(t, testCase.expected, actual)
			require.Equal(t, jsontypes.NewNormalizedValue(testCase.expected), NewNormalizedJSONValue([]byte(testCase.input)))
		})
	}
}
