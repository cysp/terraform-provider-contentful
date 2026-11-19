package provider

import (
	"slices"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *AppInstallationModel) ToXContentfulMarketplaceHeaderValue() (cm.OptString, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := cm.OptString{}

	marketplaceStrings, marketplaceStringDiags := knownOptionalStringSetElements(path.Root("marketplace"), model.Marketplace)
	diags.Append(marketplaceStringDiags...)

	if len(marketplaceStrings) > 0 {
		slices.Sort(marketplaceStrings)

		value.SetTo(strings.Join(marketplaceStrings, ","))
	}

	return value, diags
}

func (model *AppInstallationModel) ToAppInstallationData() (cm.AppInstallationData, diag.Diagnostics) {
	parameters, diags := encodeInstallationParameters(path.Root("parameters"), model.Parameters)

	return cm.AppInstallationData{Parameters: parameters}, diags
}
