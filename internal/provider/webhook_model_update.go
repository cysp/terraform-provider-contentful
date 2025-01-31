package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *WebhookModel) ToUpdateWebhookDefinitionReq(ctx context.Context) (contentfulManagement.UpdateWebhookDefinitionReq, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	req := contentfulManagement.UpdateWebhookDefinitionReq{
		Name:              model.Name.ValueString(),
		URL:               model.Url.ValueString(),
		Active:            util.BoolValueToOptBool(model.Active),
		HttpBasicUsername: contentfulManagement.NewOptNilPointerString(model.HttpBasicUsername.ValueStringPointer()),
		HttpBasicPassword: contentfulManagement.NewOptNilPointerString(model.HttpBasicPassword.ValueStringPointer()),
	}

	topics := make([]string, len(model.Topics.Elements()))
	diags.Append(model.Topics.ElementsAs(ctx, &topics, false)...)

	req.Topics = topics

	// filters:
	//   type: array
	//   items: {}

	headersList, headersListDiags := ToWebhookDefinitionHeaders(ctx, path.Root("headers"), model.Headers)
	diags.Append(headersListDiags...)

	req.Headers = headersList

	return req, diags
}
