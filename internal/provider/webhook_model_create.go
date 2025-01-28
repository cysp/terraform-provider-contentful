package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	webhookfilter "github.com/cysp/terraform-provider-contentful/internal/provider/webhook_filter"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *WebhookModel) ToCreateWebhookDefinitionReq(ctx context.Context) (contentfulManagement.CreateWebhookDefinitionReq, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	req := contentfulManagement.CreateWebhookDefinitionReq{
		Name:              model.Name.ValueString(),
		URL:               model.Url.ValueString(),
		Active:            util.BoolValueToOptBool(model.Active),
		HttpBasicUsername: contentfulManagement.NewOptNilPointerString(model.HttpBasicUsername.ValueStringPointer()),
		HttpBasicPassword: contentfulManagement.NewOptNilPointerString(model.HttpBasicPassword.ValueStringPointer()),
	}

	if model.Topics.IsNull() {
		req.Topics = nil
	} else {
		topics := make([]string, len(model.Topics.Elements()))
		diags.Append(model.Topics.ElementsAs(ctx, &topics, false)...)

		req.Topics = topics
	}

	if model.Filters.IsNull() {
		req.Filters = contentfulManagement.NewOptNilWebhookDefinitionFilterArrayNull()
	} else {
		modelFilters := make([]webhookfilter.WebhookFilterValue, len(model.Filters.Elements()))
		diags.Append(model.Filters.ElementsAs(ctx, &modelFilters, false)...)

		filters := make([]contentfulManagement.WebhookDefinitionFilter, len(modelFilters))
		for i, modelFilter := range modelFilters {
			filter, filterDiags := ToWebhookDefinitionFilter(ctx, modelFilter)
			diags.Append(filterDiags...)
			filters[i] = filter
		}
		req.Filters = contentfulManagement.NewOptNilWebhookDefinitionFilterArray(filters)
	}

	// filters:
	//   type: array
	//   items: {}

	headersList, headersListDiags := ToWebhookDefinitionHeaders(ctx, path.Root("headers"), model.Headers)
	diags.Append(headersListDiags...)

	req.Headers = headersList

	return req, diags
}
