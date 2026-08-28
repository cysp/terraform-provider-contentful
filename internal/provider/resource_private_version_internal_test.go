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

func TestRequiredPrivateVersionRequiresPositiveValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value    []byte
		want     int
		hasError bool
	}{
		"positive": {value: []byte("3"), want: 3},
		"zero":     {value: []byte("0"), hasError: true},
		"negative": {value: []byte("-1"), want: -1, hasError: true},
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
