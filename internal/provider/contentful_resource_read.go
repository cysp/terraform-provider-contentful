package provider

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func RemoveContentfulResourceIfNotFound(ctx context.Context, state *tfsdk.State, response any, err error, warningSummary string) (bool, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	responseStatus, ok := contentfulResponseStatusCode(response)
	if !ok || responseStatus != http.StatusNotFound {
		return false, diags
	}

	diags.AddWarning(warningSummary, util.ErrorDetailFromContentfulManagementResponse(response, err))
	state.RemoveResource(ctx)

	return true, diags
}

func contentfulResponseStatusCode(response any) (int, bool) {
	responseWithStatusCode, ok := response.(cm.StatusCodeResponse)
	if !ok {
		return 0, false
	}

	return responseWithStatusCode.GetStatusCode(), true
}
