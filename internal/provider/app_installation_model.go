package provider

import (
	"context"
	"slices"
	"strings"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (model *AppInstallationModel) ToXContentfulMarketplaceHeaderValue(ctx context.Context) (contentfulManagement.OptString, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := contentfulManagement.OptString{}

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

func (model *AppInstallationModel) ToPutAppInstallationReq() (contentfulManagement.PutAppInstallationReq, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	req := contentfulManagement.PutAppInstallationReq{}

	switch {
	case model.Parameters.IsUnknown():
		diags.AddAttributeWarning(path.Root("parameters"), "Failed to update app installation parameters", "Parameters are unknown")
	case model.Parameters.IsNull():
	default:
		req.Parameters = []byte(model.Parameters.ValueString())
	}

	return req, diags
}

func (model *AppInstallationModel) ReadFromResponse(appInstallation *contentfulManagement.AppInstallation) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// SpaceId, EnvironmentId and AppDefinitionId are all already known

	if appInstallation.Parameters != nil {
		constraint, err := util.JxNormalizeOpaqueBytes(appInstallation.Parameters, util.JxEncodeOpaqueOptions{EscapeStrings: true})
		if err != nil {
			diags.AddAttributeError(path.Root("parameters"), "Failed to read parameters", err.Error())
		}

		model.Parameters = jsontypes.NewNormalizedValue(string(constraint))
	} else {
		model.Parameters = jsontypes.NewNormalizedNull()
	}

	return diags
}
