package util

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type ProviderPrivateData interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
}

func PrivateDataSetValue[T interface{}](ctx context.Context, providerData ProviderPrivateData, key string, value T) diag.Diagnostics {
	diags := diag.Diagnostics{}

	valueBytes, err := json.Marshal(value)
	if err != nil {
		diags.AddError("Failed to marshal value", err.Error())
	}

	if diags.HasError() {
		return diags
	}

	return providerData.SetKey(ctx, key, valueBytes)
}

func PrivateDataGetValue[T interface{}](ctx context.Context, providerData ProviderPrivateData, key string, value *T) diag.Diagnostics {
	diags := diag.Diagnostics{}

	valueBytes, getDiags := providerData.GetKey(ctx, key)
	diags.Append(getDiags...)

	if diags.HasError() {
		return diags
	}

	err := json.Unmarshal(valueBytes, value)
	if err != nil {
		diags.AddError("Failed to unmarshal value", err.Error())
	}

	return diags
}
