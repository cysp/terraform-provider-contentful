package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (model *WebhookModel) ToCreateWebhookDefinitionReq(_ context.Context) (contentfulManagement.CreateWebhookDefinitionReq, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	req := contentfulManagement.CreateWebhookDefinitionReq{
		Name: model.Name.ValueString(),
	}

	// scopes := make([]string, len(model.Scopes.Elements()))
	// diags.Append(model.Scopes.ElementsAs(ctx, &scopes, false)...)

	// req.Scopes = scopes

	// if !model.ExpiresIn.IsNull() && !model.ExpiresIn.IsUnknown() {
	// 	req.ExpiresIn = contentfulManagement.NewOptNilPointerInt64(model.ExpiresIn.ValueInt64Pointer())
	// }

	return req, diags
}

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

	headersList, headersListDiags := ToWebhookDefinitionHeaders(ctx, path.Root("headers"), model.Headers)
	diags.Append(headersListDiags...)

	req.Headers = headersList

	return req, diags
}

func (model *WebhookModel) ReadFromResponse(ctx context.Context, webhookDefinition *contentfulManagement.WebhookDefinition) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model.WebhookId = types.StringValue(webhookDefinition.Sys.ID)

	model.Name = types.StringValue(webhookDefinition.Name)

	model.Url = types.StringValue(webhookDefinition.URL)

	topicsList, topicsListDiags := types.ListValueFrom(ctx, types.String{}.Type(ctx), webhookDefinition.Topics)
	diags.Append(topicsListDiags...)

	model.Topics = topicsList

	if httpBasicUsername, ok := webhookDefinition.HttpBasicUsername.Get(); ok {
		model.HttpBasicUsername = types.StringValue(httpBasicUsername)
	} else {
		model.HttpBasicUsername = types.StringNull()
	}

	if httpBasicPassword, ok := webhookDefinition.HttpBasicPassword.Get(); ok {
		model.HttpBasicPassword = types.StringValue(httpBasicPassword)
	} else {
		model.HttpBasicPassword = types.StringNull()
	}

	headersList, headersListDiags := ReadHeadersListValueFromResponse(ctx, path.Root("headers"), model.Headers, webhookDefinition.Headers)
	diags.Append(headersListDiags...)

	model.Headers = headersList

	// filters:
	//   type: array
	//   items: {}

	// transformation:
	//   type: object
	//   properties:
	//     method:
	//       type: string
	//       nullable: true
	//     contentType:
	//       type: string
	//       nullable: true
	//     includeContentLength:
	//       type: boolean
	//       nullable: true
	//     body: {}
	//   nullable: true

	model.Active = util.OptBoolToBoolValue(webhookDefinition.Active)

	return diags
}
