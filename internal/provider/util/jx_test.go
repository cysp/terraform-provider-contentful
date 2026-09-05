package util_test

import (
	"encoding/json"
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/go-faster/jx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONMarshalEscaping(t *testing.T) {
	t.Parallel()

	input := map[string]string{
		"a<b&c>d": "e<f&g>h",
	}

	expected := []byte(`{"a\u003cb\u0026c\u003ed":"e\u003cf\u0026g\u003eh"}`)

	actual, err := json.Marshal(input)

	assert.Equal(t, expected, actual)
	assert.NoError(t, err)
}

func TestJxNormalizeOpaqueBytes(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input                      []byte
		expected                   []byte
		expectedWithEscapedStrings []byte
	}{
		"object": {
			input:                      []byte(`{"string": "string", "number": 2, "boolean": true, "null": null, "string_with_escapes": "a<b&c>d"}`),
			expected:                   []byte(`{"boolean":true,"null":null,"number":2,"string":"string","string_with_escapes":"a<b&c>d"}`),
			expectedWithEscapedStrings: []byte(`{"boolean":true,"null":null,"number":2,"string":"string","string_with_escapes":"a\u003cb\u0026c\u003ed"}`),
		},
		"object key escaping is independent of string value escaping": {
			input:                      []byte(`{"a<b&c>d\"\\\n": "e<f&g>h"}`),
			expected:                   []byte(`{"a<b&c>d\"\\\n":"e<f&g>h"}`),
			expectedWithEscapedStrings: []byte(`{"a<b&c>d\"\\\n":"e\u003cf\u0026g\u003eh"}`),
		},
		"nested objects and arrays with escaping": {
			input:                      []byte(`{"z": [{"z": "<last>", "a&b": "first&"}, "<second>"], "a": {"<key>": ">value"}}`),
			expected:                   []byte(`{"a":{"<key>":">value"},"z":[{"a&b":"first&","z":"<last>"},"<second>"]}`),
			expectedWithEscapedStrings: []byte(`{"a":{"<key>":"\u003evalue"},"z":[{"a&b":"first\u0026","z":"\u003clast\u003e"},"\u003csecond\u003e"]}`),
		},
		"array": {
			input:    []byte(`["string", 2, true]`),
			expected: []byte(`["string",2,true]`),
		},
		"string": {
			input:    []byte(`"string"`),
			expected: []byte(`"string"`),
		},
		"number": {
			input:    []byte(`123`),
			expected: []byte(`123`),
		},
		"integer beyond float64 exact range": {
			input:    []byte(`9007199254740993`),
			expected: []byte(`9007199254740993`),
		},
		"high precision decimal": {
			input:    []byte(`0.12345678901234567890123456789`),
			expected: []byte(`0.12345678901234567890123456789`),
		},
		"exponent notation": {
			input:    []byte(`1.234567890123456789e+100`),
			expected: []byte(`1.234567890123456789e+100`),
		},
		"boolean": {
			input:    []byte(`true`),
			expected: []byte(`true`),
		},
		"null": {
			input:    []byte(`null`),
			expected: []byte(`null`),
		},
	}

	for name, testcase := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := util.JxNormalizeOpaqueBytes(testcase.input, util.JxEncodeOpaqueOptions{EscapeStrings: false})

			assert.Equal(t, testcase.expected, actual)
			assert.NoError(t, err)
		})

		t.Run(name+" with escaped strings", func(t *testing.T) {
			t.Parallel()

			actual, err := util.JxNormalizeOpaqueBytes(testcase.input, util.JxEncodeOpaqueOptions{EscapeStrings: true})

			expected := testcase.expectedWithEscapedStrings
			if expected == nil {
				expected = testcase.expected
			}

			assert.Equal(t, expected, actual)
			require.NoError(t, err)
			assert.True(t, json.Valid(actual), "normalized output must be valid JSON")
		})
	}
}

func TestJxNormalizeOpaqueBytesInvalid(t *testing.T) {
	t.Parallel()

	tests := map[string][]byte{
		"invalid value":               []byte("invalid"),
		"trailing value":              []byte("true false"),
		"trailing non-value":          []byte("9007199254740993 trailing"),
		"trailing object":             []byte(`{}{}`),
		"trailing malformed value":    []byte(`true {`),
		"invalid nested object value": []byte(`{"a": {"b": invalid}}`),
		"invalid nested array value":  []byte(`{"a": [true, invalid]}`),
		"incomplete nested structure": []byte(`{"a": [{"b": true}`),
	}

	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := util.JxNormalizeOpaqueBytes(input, util.JxEncodeOpaqueOptions{})

			assert.Error(t, err)
		})
	}
}

func TestJxDecodeOpaque(t *testing.T) {
	t.Parallel()

	input := []byte(`{"a": "b", "c": 4, "d": true}`)
	expected := map[string]any{
		"a": "b",
		"c": json.Number("4"),
		"d": true,
	}

	decoder := jx.DecodeBytes(input)
	actual, err := util.JxDecodeOpaque(decoder)

	assert.Equal(t, expected, actual)
	assert.NoError(t, err)
}

func TestJxEncodeOpaqueOrdered(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"object": map[string]any{
			"b": 1,
			"d": 2,
			"c": 3,
			"a": 4,
		},
		"array": []any{
			"b",
			"d",
			"c",
			"a",
		},
		"string": "abcd",
		"int":    int(-42),
		"int8":   []any{int8(-128), int8(127)},
		"int16":  []any{int16(-32768), int16(32767)},
		"int32":  []any{int32(-2147483648), int32(2147483647)},
		"int64":  []any{int64(-9223372036854775808), int64(9223372036854775807)},
		"uint":   uint(42),
		"uint8":  []any{uint8(0), uint8(255)},
		"uint16": []any{uint16(0), uint16(65535)},
		"uint32": []any{uint32(0), uint32(4294967295)},
		"uint64": []any{uint64(0), uint64(18446744073709551615)},
		"float32": []any{
			float32(-3.40282346638528859811704183484516925440e+38),
			float32(3.40282346638528859811704183484516925440e+38),
		},
		"float64": []any{
			float64(-1.797693134862315708145274237317043567981e+308),
			float64(1.797693134862315708145274237317043567981e+308),
		},
		"bool": []any{true, false},
		"null": []any{nil},
		"number": []any{
			json.Number("-0"),
			json.Number("9007199254740993"),
			json.Number("1.234567890123456789e+100"),
		},
	}
	expected := `{` +
		`"array":["b","d","c","a"],` +
		`"bool":[true,false],` +
		`"float32":[-3.4028235e+38,3.4028235e+38],` +
		`"float64":[-1.7976931348623157e+308,1.7976931348623157e+308],` +
		`"int":-42,` +
		`"int16":[-32768,32767],` +
		`"int32":[-2147483648,2147483647],` +
		`"int64":[-9223372036854775808,9223372036854775807],` +
		`"int8":[-128,127],` +
		`"null":[null],` +
		`"number":[-0,9007199254740993,1.234567890123456789e+100],` +
		`"object":{"a":4,"b":1,"c":3,"d":2},` +
		`"string":"abcd",` +
		`"uint":42,` +
		`"uint16":[0,65535],` +
		`"uint32":[0,4294967295],` +
		`"uint64":[0,18446744073709551615],` +
		`"uint8":[0,255]` +
		`}`

	encoder := jx.Encoder{}
	err := util.JxEncodeOpaqueOrdered(&encoder, input, util.JxEncodeOpaqueOptions{EscapeStrings: false})

	assert.Equal(t, expected, string(encoder.Bytes()))
	assert.NoError(t, err)
}

func TestJxEncodeOpaqueOrderedInvalid(t *testing.T) {
	t.Parallel()

	type Unencodable struct{}

	tests := map[string]struct {
		input any
	}{
		"root": {
			input: Unencodable{},
		},
		"object": {
			input: map[string]any{
				"key": Unencodable{},
			},
		},
		"array": {input: []any{
			Unencodable{},
		}},
		"invalid number": {
			input: json.Number("not-a-number"),
		},
		"empty number": {
			input: json.Number(""),
		},
	}

	for name, testcase := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			encoder := jx.Encoder{}
			err := util.JxEncodeOpaqueOrdered(&encoder, testcase.input, util.JxEncodeOpaqueOptions{})

			assert.Error(t, err)
		})
	}
}
