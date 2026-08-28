package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
)

type privateVersionDataStub struct {
	value []byte
}

func (s *privateVersionDataStub) GetKey(context.Context, string) ([]byte, diag.Diagnostics) {
	return s.value, nil
}

func (*privateVersionDataStub) SetKey(context.Context, string, []byte) diag.Diagnostics {
	return nil
}

func TestRequiredPrivateVersionRequiresReadableValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value    []byte
		want     int
		hasError bool
	}{
		"positive": {value: []byte("3"), want: 3},
		"zero":     {value: []byte("0")},
		"negative": {value: []byte("-1"), want: -1},
		"null":     {value: []byte("null")},
		"missing":  {hasError: true},
		"malformed": {
			value:    []byte(`"invalid"`),
			hasError: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			version, diags := requiredPrivateVersion(t.Context(), &privateVersionDataStub{value: test.value})
			assert.Equal(t, test.want, version)
			assert.Equal(t, test.hasError, diags.HasError())
		})
	}
}

func TestOptionalPrivateVersionDistinguishesAbsenceFromDecodedValues(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value    []byte
		want     int
		found    bool
		hasError bool
	}{
		"positive":  {value: []byte("3"), want: 3, found: true},
		"zero":      {value: []byte("0"), found: true},
		"negative":  {value: []byte("-1"), want: -1, found: true},
		"null":      {value: []byte("null"), found: true},
		"missing":   {},
		"malformed": {value: []byte(`"invalid"`), found: true, hasError: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			version, found, diags := optionalPrivateVersion(t.Context(), &privateVersionDataStub{value: test.value})
			assert.Equal(t, test.want, version)
			assert.Equal(t, test.found, found)
			assert.Equal(t, test.hasError, diags.HasError())
		})
	}
}
