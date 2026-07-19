package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func (model *WebhookModel) ToWebhookDefinitionData(ctx context.Context, path path.Path) (cm.WebhookDefinitionData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	name, nameDiags := KnownStringValue(model.Name, path.AtName("name"))
	diags.Append(nameDiags...)

	url, urlDiags := KnownStringValue(model.URL, path.AtName("url"))
	diags.Append(urlDiags...)

	active, activeDiags := KnownBoolValue(model.Active, path.AtName("active"))
	diags.Append(activeDiags...)

	req := cm.WebhookDefinitionData{
		Name:   name,
		URL:    url,
		Active: cm.NewOptBool(active),
	}

	if model.HTTPBasicUsername.IsUnknown() {
		diags.AddAttributeError(path.AtName("http_basic_username"), "Unexpected unknown HTTP basic username", "The HTTP basic username must be known before it can be sent to Contentful.")
	} else {
		req.HttpBasicUsername = cm.NewOptNilPointerString(model.HTTPBasicUsername.ValueStringPointer())
	}

	if model.HTTPBasicPassword.IsUnknown() {
		diags.AddAttributeError(path.AtName("http_basic_password"), "Unexpected unknown HTTP basic password", "The HTTP basic password must be known before it can be sent to Contentful.")
	} else {
		req.HttpBasicPassword = cm.NewOptNilPointerString(model.HTTPBasicPassword.ValueStringPointer())
	}

	switch {
	case model.Topics.IsNull():
		req.Topics = nil
	case model.Topics.IsUnknown():
		diags.AddAttributeError(path.AtName("topics"), "Unexpected unknown webhook topics", "Webhook topics must be known before they can be sent to Contentful.")
	default:
		topics := make([]string, len(model.Topics.Elements()))
		diags.Append(tfsdk.ValueAs(ctx, model.Topics, &topics)...)

		req.Topics = topics
	}

	filters, filtersDiags := ToOptNilWebhookDefinitionFilterArray(ctx, path.AtName("filters"), model.Filters)
	diags.Append(filtersDiags...)

	req.Filters = filters

	headersList, headersListDiags := ToWebhookDefinitionHeaders(ctx, path.AtName("headers"), model.Headers)
	diags.Append(headersListDiags...)

	req.Headers = headersList

	transformation, transformationDiags := ToOptNilWebhookDefinitionDataTransformation(ctx, path.AtName("transformation"), model.Transformation)
	diags.Append(transformationDiags...)

	req.Transformation = transformation

	return req, diags
}
