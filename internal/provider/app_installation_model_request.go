package provider

import (
	"context"
	"slices"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *AppInstallationModel) ToXContentfulMarketplaceHeaderValue(ctx context.Context) (cm.OptString, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := cm.OptString{}

	marketplaceStrings, marketplaceStringDiags := model.ToXContentfulMarketplaceHeaderValueElements(ctx)
	diags.Append(marketplaceStringDiags...)

	if len(marketplaceStrings) > 0 {
		slices.Sort(marketplaceStrings)

		value.SetTo(strings.Join(marketplaceStrings, ","))
	}

	return value, diags
}

func (model *AppInstallationModel) ToXContentfulMarketplaceHeaderValueElements(_ context.Context) ([]string, diag.Diagnostics) {
	return knownOptionalStringSetElements(path.Root("marketplace"), model.Marketplace)
}

func (model *AppInstallationModel) ToAppInstallationData() (cm.AppInstallationData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fields := cm.AppInstallationData{}

	switch {
	case model.Parameters.IsUnknown():
		diags.AddAttributeError(
			path.Root("parameters"),
			"Unexpected unknown app installation parameters",
			"App installation parameters must be known before they can be sent to Contentful.",
		)

		return cm.AppInstallationData{}, diags
	case model.Parameters.IsNull():
	default:
		fields.Parameters = []byte(model.Parameters.ValueString())
	}

	return fields, diags
}
