package util_test

import (
	"context"
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
)

type providerPrivateData struct {
	data map[string][]byte
}

func newProviderPrivateData() *providerPrivateData {
	return &providerPrivateData{
		data: make(map[string][]byte),
	}
}

func (p *providerPrivateData) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value, found := p.data[key]
	if !found {
		diags.AddError("Private Data: key not found", "Key not found: "+key)
	}

	return value, diags
}

func (p *providerPrivateData) SetKey(_ context.Context, key string, value []byte) diag.Diagnostics {
	diags := diag.Diagnostics{}

	p.data[key] = value

	return diags
}

var _ util.ProviderPrivateData = &providerPrivateData{}

func TestPrivateDataGetIntNotSet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	privateData := newProviderPrivateData()

	value, diags := util.PrivateDataGetInt(ctx, privateData, "key")

	assert.EqualValues(t, 0, value)
	assert.NotEmpty(t, diags)
}

func TestPrivateDataGetSetInt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	privateData := newProviderPrivateData()

	diags := util.PrivateDataSetInt(ctx, privateData, "key", 42)

	assert.EqualValues(t, []byte{'4', '2'}, privateData.data["key"])
	assert.Empty(t, diags)

	value, diags := util.PrivateDataGetInt(ctx, privateData, "key")

	assert.EqualValues(t, 42, value)
	assert.Empty(t, diags)
}

func TestPrivateDataGetIntInvalid(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	privateData := newProviderPrivateData()
	privateData.data["key"] = []byte("invalid")

	value, diags := util.PrivateDataGetInt(ctx, privateData, "key")

	assert.EqualValues(t, 0, value)
	assert.NotEmpty(t, diags)
}
