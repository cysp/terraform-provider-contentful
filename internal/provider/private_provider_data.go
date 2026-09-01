package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type PrivateProviderData interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
}

func SetPrivateProviderData[T any](ctx context.Context, providerData PrivateProviderData, key string, value T) diag.Diagnostics {
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

func requiredPrivateVersion(ctx context.Context, providerData PrivateProviderData) (int, diag.Diagnostics) {
	version, found, diags := optionalPrivateVersion(ctx, providerData)
	if !found && !diags.HasError() {
		diags.AddError("Private version is unavailable", "Terraform private state does not contain a Contentful version.")
	}

	return version, diags
}

func optionalPrivateVersion(ctx context.Context, providerData PrivateProviderData) (int, bool, diag.Diagnostics) {
	value, diags := providerData.GetKey(ctx, "version")
	if diags.HasError() || len(value) == 0 {
		return 0, false, diags
	}

	var version int

	err := json.Unmarshal(value, &version)
	if err != nil {
		diags.AddError("Failed to unmarshal value", err.Error())

		return 0, true, diags
	}

	return version, true, diags
}
