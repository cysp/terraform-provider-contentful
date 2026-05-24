package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const contentfulResourceVersionPrivateStateKey = "version"

func SetContentfulResourceVersion(ctx context.Context, providerData PrivateProviderData, version int) diag.Diagnostics {
	return SetPrivateProviderData(ctx, providerData, contentfulResourceVersionPrivateStateKey, version)
}

func GetContentfulResourceVersion(ctx context.Context, providerData PrivateProviderData) (int, diag.Diagnostics) {
	var version int

	diags := GetPrivateProviderData(ctx, providerData, contentfulResourceVersionPrivateStateKey, &version)

	return version, diags
}
