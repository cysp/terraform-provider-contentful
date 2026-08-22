package provider

import (
	"context"
	"encoding/json"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const taxonomyVersionPrivateKey = "version"

func taxonomyPriorStateVersion(ctx context.Context, privateData PrivateProviderData) (int, diag.Diagnostics) {
	version, found, diags := taxonomyPrivateVersion(ctx, privateData)
	if diags.HasError() {
		return 0, diags
	}

	if !found {
		diags.AddError("Taxonomy resource version is unavailable", "Contentful sys.version was not captured in Terraform private state. Refresh the Terraform state and create a new plan before applying this change.")

		return 0, diags
	}

	return version, diags
}

func taxonomyPrivateVersion(ctx context.Context, privateData PrivateProviderData) (int, bool, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	encoded, privateDiags := privateData.GetKey(ctx, taxonomyVersionPrivateKey)
	diags.Append(privateDiags...)

	if diags.HasError() {
		return 0, false, diags
	}

	if len(encoded) == 0 {
		return 0, false, nil
	}

	version := 0

	err := json.Unmarshal(encoded, &version)
	if err != nil {
		diags.AddError("Invalid taxonomy resource version", "Contentful sys.version captured in Terraform private state is not a valid integer. Refresh the Terraform state and create a new plan before applying this change.\n\n"+err.Error())

		return 0, true, diags
	}

	if version <= 0 {
		diags.AddError("Invalid taxonomy resource version", "Contentful sys.version captured in Terraform private state is not positive. Refresh the Terraform state and create a new plan before applying this change.")

		return 0, true, diags
	}

	return version, true, diags
}

func taxonomyDeleteResponseVersion(resourceName, expectedOrganizationID, expectedResourceID, actualOrganizationID, actualResourceID string, version int) (int, diag.Diagnostics) {
	if actualOrganizationID != expectedOrganizationID || actualResourceID != expectedResourceID {
		return 0, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Unexpected Contentful "+resourceName+" response",
			"The response identity differed from the requested taxonomy endpoint.",
		)}
	}

	return taxonomyResponseVersion(version)
}

func taxonomyResponseVersion(version int) (int, diag.Diagnostics) {
	if version <= 0 {
		return 0, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Invalid taxonomy resource version",
			"Contentful returned a nonpositive sys.version, which cannot be stored in Terraform private state.",
		)}
	}

	return version, nil
}

func setTaxonomyPrivateVersion(ctx context.Context, privateData PrivateProviderData, version int) diag.Diagnostics {
	return SetPrivateProviderData(ctx, privateData, taxonomyVersionPrivateKey, version)
}

func taxonomyMutationErrorDetail(response any, err error) string {
	if taxonomyResponseIsVersionMismatch(response) {
		return "Contentful rejected the change because the taxonomy resource changed after this Terraform plan was created. Refresh the Terraform state and create a new plan before applying again.\n\n" +
			util.ErrorDetailFromContentfulManagementResponse(response, err)
	}

	return util.ErrorDetailFromContentfulManagementResponse(response, err)
}

func taxonomyResponseIsVersionMismatch(response any) bool {
	errorResponse, ok := response.(cm.ErrorStatusCodeResponse)
	if !ok {
		return false
	}

	contentfulError, ok := errorResponse.GetError()

	return ok && contentfulError.Sys.ID == cm.ErrorSysIDVersionMismatch
}
