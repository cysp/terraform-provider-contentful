package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *WebhookModel) ToWebhookDefinitionData(ctx context.Context, path path.Path) (cm.WebhookDefinitionData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	name, nameDiags := requestRequiredString(model.Name, path.AtName("name"))
	diags.Append(nameDiags...)

	url, urlDiags := requestRequiredString(model.URL, path.AtName("url"))
	diags.Append(urlDiags...)

	active, activeDiags := requestRequiredBool(model.Active, path.AtName("active"))
	diags.Append(activeDiags...)

	httpBasicUsername, httpBasicUsernameDiags := webhookOptionalNullableString(model.HTTPBasicUsername, path.AtName("http_basic_username"))
	diags.Append(httpBasicUsernameDiags...)

	httpBasicPassword, httpBasicPasswordDiags := webhookOptionalNullableString(model.HTTPBasicPassword, path.AtName("http_basic_password"))
	diags.Append(httpBasicPasswordDiags...)

	req := cm.WebhookDefinitionData{
		Name:              name,
		URL:               url,
		Active:            cm.NewOptBool(active),
		HttpBasicUsername: httpBasicUsername,
		HttpBasicPassword: httpBasicPassword,
	}

	switch {
	case model.Topics.IsUnknown():
		diags.AddAttributeError(
			path.AtName("topics"),
			"Unexpected unknown webhook topics",
			"Webhook topics must be known before they can be sent to Contentful.",
		)
	case model.Topics.IsNull():
		req.Topics = nil
	default:
		topics, topicDiags := knownStringListElements(path.AtName("topics"), model.Topics.Elements())
		diags.Append(topicDiags...)

		req.Topics = topics
	}

	filters, filterDiags := ToOptNilWebhookDefinitionFilterArray(ctx, path.AtName("filters"), model.Filters)
	diags.Append(filterDiags...)

	req.Filters = filters

	headersList, headersListDiags := ToWebhookDefinitionHeaders(ctx, path.AtName("headers"), model.Headers)
	diags.Append(headersListDiags...)

	req.Headers = headersList

	transformation, transformationDiags := ToOptNilWebhookDefinitionDataTransformation(ctx, path.AtName("transformation"), model.Transformation)
	diags.Append(transformationDiags...)

	req.Transformation = transformation

	if diags.HasError() {
		return cm.WebhookDefinitionData{}, diags
	}

	return req, diags
}
