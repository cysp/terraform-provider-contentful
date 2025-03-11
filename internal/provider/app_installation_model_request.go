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

func (model *AppInstallationResourceModel) ToXContentfulMarketplaceHeaderValue(ctx context.Context) (cm.OptString, diag.Diagnostics) {
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

func (model *AppInstallationResourceModel) ToXContentfulMarketplaceHeaderValueElements(ctx context.Context) ([]string, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if model.Marketplace.IsNull() || model.Marketplace.IsUnknown() {
		return []string{}, diags
	}

	marketplaceElements := make([]types.String, len(model.Marketplace.Elements()))
	diags.Append(model.Marketplace.ElementsAs(ctx, &marketplaceElements, false)...)

	marketplaceStrings := make([]string, 0, len(marketplaceElements))

	for _, element := range marketplaceElements {
		if !element.IsNull() && !element.IsUnknown() {
			marketplaceStrings = append(marketplaceStrings, element.ValueString())
		}
	}

	return marketplaceStrings, diags
}

func (model *AppInstallationResourceModel) ToAppInstallationFields() (cm.AppInstallationFields, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fields := cm.AppInstallationFields{}

	switch {
	case model.Parameters.IsUnknown():
		diags.AddAttributeWarning(path.Root("parameters"), "Failed to update app installation parameters", "Parameters are unknown")
	case model.Parameters.IsNull():
	default:
		fields.Parameters = []byte(model.Parameters.ValueString())
	}

	return fields, diags
}
