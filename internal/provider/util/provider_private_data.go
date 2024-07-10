package util

import (
	"context"

	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type ProviderPrivateData interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
}

func PrivateDataSetInt(ctx context.Context, providerData ProviderPrivateData, key string, value int) diag.Diagnostics {
	encoder := jx.Encoder{}
	encoder.Int(value)
	valueBytes := encoder.Bytes()

	return providerData.SetKey(ctx, key, valueBytes)
}

func PrivateDataGetInt(ctx context.Context, providerData ProviderPrivateData, key string) (int, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	valueBytes, getDiags := providerData.GetKey(ctx, key)
	diags.Append(getDiags...)

	if diags.HasError() {
		return 0, diags
	}

	decoder := jx.DecodeBytes(valueBytes)

	value, err := decoder.Int()
	if err != nil {
		diags.AddError("Failed to decode int", err.Error())
	}

	return value, diags
}
