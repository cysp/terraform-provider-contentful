package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func SetProviderDataFromDataSourceConfigureRequest[ProviderData any](req datasource.ConfigureRequest, out *ProviderData) diag.Diagnostics {
	return setProviderData(req.ProviderData, out)
}

func SetProviderDataFromResourceConfigureRequest[ProviderData any](req resource.ConfigureRequest, out *ProviderData) diag.Diagnostics {
	return setProviderData(req.ProviderData, out)
}

func setProviderData[ProviderData any](data any, out *ProviderData) diag.Diagnostics {
	if data == nil {
		return nil
	}

	providerData, ok := data.(ProviderData)
	if !ok {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Invalid provider data", "")}
	}

	*out = providerData

	return nil
}
