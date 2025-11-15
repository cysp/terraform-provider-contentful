package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (m *ResourceProviderModel) ToResourceProviderRequest(_ context.Context, _ path.Path) (cm.ResourceProviderRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	req := cm.ResourceProviderRequest{
		Sys:      cm.NewResourceProviderRequestSys(m.ResourceProviderID.ValueString()),
		Type:     cm.ResourceProviderRequestTypeFunction,
		Function: cm.NewFunctionLink(m.FunctionID.ValueString()),
	}

	return req, diags
}
