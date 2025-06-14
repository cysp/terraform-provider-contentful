package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (m *AppDefinitionResourceProviderModel) ToResourceProviderRequest(_ context.Context, _ path.Path) (cm.ResourceProviderRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	req := cm.ResourceProviderRequest{
		Sys: cm.ResourceProviderRequestSys{
			ID: m.ResourceProviderID.ValueString(),
		},
		Type: cm.ResourceProviderRequestTypeFunction,
		Function: cm.FunctionLink{
			Sys: cm.FunctionLinkSys{
				Type:     cm.FunctionLinkSysTypeLink,
				LinkType: cm.FunctionLinkSysLinkTypeFunction,
				ID:       m.FunctionID.ValueString(),
			},
		},
	}

	return req, diags
}
