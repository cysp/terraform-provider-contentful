package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (model *WebhookModel) ReadFromResponse(ctx context.Context, webhookDefinition *contentfulManagement.WebhookDefinition) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model.WebhookId = types.StringValue(webhookDefinition.Sys.ID)

	model.Name = types.StringValue(webhookDefinition.Name)

	model.Url = types.StringValue(webhookDefinition.URL)

	topicsList, topicsListDiags := types.ListValueFrom(ctx, types.String{}.Type(ctx), webhookDefinition.Topics)
	diags.Append(topicsListDiags...)

	model.Topics = topicsList

	// filters:
	//   type: array
	//   items: {}

	model.HttpBasicUsername = types.StringPointerValue(webhookDefinition.HttpBasicUsername.ValueStringPointer())
	model.HttpBasicPassword = types.StringPointerValue(webhookDefinition.HttpBasicPassword.ValueStringPointer())

	headersList, headersListDiags := ReadHeadersListValueFromResponse(ctx, path.Root("headers"), model.Headers, webhookDefinition.Headers)
	diags.Append(headersListDiags...)

	model.Headers = headersList

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
