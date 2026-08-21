package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (m *ResourceProviderModel) ToResourceProviderRequest(_ context.Context, modelPath path.Path) (cm.ResourceProviderRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	resourceProviderID, resourceProviderIDDiags := requestRequiredString(m.ResourceProviderID, modelPath.AtName("resource_provider_id"))
	diags.Append(resourceProviderIDDiags...)

	functionID, functionIDDiags := requestRequiredString(m.FunctionID, modelPath.AtName("function_id"))
	diags.Append(functionIDDiags...)

	if diags.HasError() {
		return cm.ResourceProviderRequest{}, diags
	}

	req := cm.ResourceProviderRequest{
		Sys:      cm.NewResourceProviderRequestSys(resourceProviderID),
		Type:     cm.ResourceProviderRequestTypeFunction,
		Function: cm.NewFunctionLink(functionID),
	}

	return req, diags
}
