package provider_test

import (
	"context"
	"math"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type privateProviderData struct {
	data map[string][]byte
}

func newProviderPrivateData() *privateProviderData {
	return &privateProviderData{
		data: make(map[string][]byte),
	}
}

func (p *privateProviderData) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value, found := p.data[key]
	if !found {
		diags.AddError("Private Data: key not found", "Key not found: "+key)
	}

	return value, diags
}

func (p *privateProviderData) SetKey(_ context.Context, key string, value []byte) diag.Diagnostics {
	diags := diag.Diagnostics{}

	p.data[key] = value

	return diags
}

var _ PrivateProviderData = &privateProviderData{}

func TestSetPrivateProviderDataWritesJSON(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	privateData := newProviderPrivateData()

	diags := SetPrivateProviderData(ctx, privateData, "key", 42)

	require.Empty(t, diags)
	assert.Equal(t, []byte("42"), privateData.data["key"])
}

func TestSetPrivateProviderDataRejectsUnsupportedJSONWithoutMutation(t *testing.T) {
	t.Parallel()

	for name, value := range map[string]float64{
		"positive infinity": math.Inf(1),
		"negative infinity": math.Inf(-1),
		"NaN":               math.NaN(),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			privateData := newProviderPrivateData()
			privateData.data["key"] = []byte("42")

			diags := SetPrivateProviderData(t.Context(), privateData, "key", value)

			require.Len(t, diags.Errors(), 1)
			assert.Equal(t, "Failed to marshal value", diags.Errors()[0].Summary())
			assert.Equal(t, []byte("42"), privateData.data["key"])
		})
	}
}
