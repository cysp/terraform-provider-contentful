package provider_test

import (
	"context"
	"math"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
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

func TestPrivateDataSetInt(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	privateData := newProviderPrivateData()

	diags := SetPrivateProviderData(ctx, privateData, "key", 42)

	assert.Equal(t, []byte{'4', '2'}, privateData.data["key"])
	assert.Empty(t, diags)
}

func TestPrivateDataSetInf(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	privateData := newProviderPrivateData()

	diags := SetPrivateProviderData(ctx, privateData, "key", math.Inf(1))

	assert.NotEmpty(t, diags)
}
