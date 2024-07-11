package util_test

import (
	"context"
	"math"
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

	var value int
	diags := util.PrivateDataGetValue(ctx, privateData, "key", &value)

	assert.EqualValues(t, 0, value)
	assert.NotEmpty(t, diags)
}

func TestPrivateDataGetSetInt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	privateData := newProviderPrivateData()

	diags := util.PrivateDataSetValue(ctx, privateData, "key", 42)

	assert.EqualValues(t, []byte{'4', '2'}, privateData.data["key"])
	assert.Empty(t, diags)

	var value int
	diags = util.PrivateDataGetValue(ctx, privateData, "key", &value)

	assert.EqualValues(t, 42, value)
	assert.Empty(t, diags)
}

func TestPrivateDataGetIntInvalid(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	privateData := newProviderPrivateData()
	privateData.data["key"] = []byte("invalid")

	var value int
	diags := util.PrivateDataGetValue(ctx, privateData, "key", &value)

	assert.EqualValues(t, 0, value)
	assert.NotEmpty(t, diags)
}

func TestPrivateDataSetInf(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	privateData := newProviderPrivateData()

	diags := util.PrivateDataSetValue(ctx, privateData, "key", math.Inf(1))

	assert.NotEmpty(t, diags)
}
