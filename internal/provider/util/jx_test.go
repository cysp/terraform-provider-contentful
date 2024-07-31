package util_test

import (
	"encoding/json"
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/go-faster/jx"
	"github.com/stretchr/testify/assert"
)

func TestJsonMarshalEscaping(t *testing.T) {
	t.Parallel()

	input := map[string]string{
		"a<b&c>d": "e<f&g>h",
	}

	expected := []byte(`{"a\u003cb\u0026c\u003ed":"e\u003cf\u0026g\u003eh"}`)

	actual, err := json.Marshal(input)

	assert.Equal(t, expected, actual)
	assert.NoError(t, err)
}

func TestEncodeJxRawMapOrdered(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    map[string]jx.Raw
		expected string
	}{
		"empty": {
			input:    map[string]jx.Raw{},
			expected: "{}",
		},
		"a=b": {
			input: map[string]jx.Raw{
				"a": jx.Raw(`"b"`),
			},
			expected: `{"a":"b"}`,
		},
		"a=2": {
			input: map[string]jx.Raw{
				"a": jx.Raw(`2`),
			},
			expected: `{"a":2}`,
		},
		"a<b&c>d=e<f&g>h": {
			input: map[string]jx.Raw{
				"a<b&c>d": jx.Raw(`"e<f&g>h"`),
			},
			// n.b. not applying escaping to keys, since i don't need that at this time
			expected: `{"a<b&c>d":"e\u003cf\u0026g\u003eh"}`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			encoder := jx.Encoder{}

			util.EncodeJxRawMapOrdered(&encoder, test.input)

			assert.Equal(t, test.expected, encoder.String())
		})
	}
}
