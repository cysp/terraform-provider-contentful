package util

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func ProviderDataFromDataSourceConfigureRequest[ProviderData interface{}](req datasource.ConfigureRequest, out *ProviderData, resp *datasource.ConfigureResponse) bool {
	if req.ProviderData == nil {
		return false
	}

	if providerData, ok := req.ProviderData.(ProviderData); ok {
		*out = providerData

		return true
	}

	resp.Diagnostics.AddError("Invalid provider data", "")

	return false
}

func ProviderDataFromResourceConfigureRequest[ProviderData interface{}](req resource.ConfigureRequest, out *ProviderData, resp *resource.ConfigureResponse) bool {
	if req.ProviderData == nil {
		return false
	}

	if providerData, ok := req.ProviderData.(ProviderData); ok {
		*out = providerData

		return true
	}

	resp.Diagnostics.AddError("Invalid provider data", "")

	return false
}
