package provider

import (
	"context"
	"slices"
	"strings"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/cysp/terraform-provider-contentful/internal/tf"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *AppInstallationModel) ToXContentfulMarketplaceHeaderValue(ctx context.Context) (contentfulManagement.OptString, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := contentfulManagement.OptString{}

	marketplaceStrings, marketplaceStringDiags := tf.ElementsAsStringSlice(ctx, model.Marketplace)
	diags.Append(marketplaceStringDiags...)

	if len(marketplaceStrings) > 0 {
		slices.Sort(marketplaceStrings)

		value.SetTo(strings.Join(marketplaceStrings, ","))
	}

	return value, diags
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
