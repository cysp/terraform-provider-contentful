package provider

import (
	"context"
	"slices"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func (model *AppInstallationModel) ToXContentfulMarketplaceHeaderValueElements(ctx context.Context) ([]string, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if model.Marketplace.IsNull() {
		return []string{}, diags
	}

	if model.Marketplace.IsUnknown() {
		diags.AddAttributeError(path.Root("marketplace"), "Unexpected unknown marketplace", "Marketplace values must be known before they can be sent to Contentful.")

		return nil, diags
	}

	marketplaceElements := make([]types.String, len(model.Marketplace.Elements()))
	diags.Append(model.Marketplace.ElementsAs(ctx, &marketplaceElements, false)...)

	marketplaceStrings := make([]string, 0, len(marketplaceElements))

	for _, element := range marketplaceElements {
		elementPath := path.Root("marketplace").AtSetValue(element)
		if element.IsNull() || element.IsUnknown() {
			// Null and unknown set elements cannot provide a stable value-based path.
			elementPath = path.Root("marketplace")
		}

		value, valueDiags := KnownStringValue(element, elementPath)
		diags.Append(valueDiags...)

		if valueDiags.HasError() {
			continue
		}

		marketplaceStrings = append(marketplaceStrings, value)
	}

	if diags.HasError() {
		return nil, diags
	}

	return marketplaceStrings, diags
}

func (model *AppInstallationModel) ToAppInstallationData() (cm.AppInstallationData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fields := cm.AppInstallationData{}

	switch {
	case model.Parameters.IsUnknown():
		diags.AddAttributeError(path.Root("parameters"), "Unexpected unknown app installation parameters", "App installation parameters must be known before they can be sent to Contentful.")
	case model.Parameters.IsNull():
	default:
		fields.Parameters = []byte(model.Parameters.ValueString())
	}

	return fields, diags
}
