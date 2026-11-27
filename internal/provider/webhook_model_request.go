package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (model *WebhookModel) ToWebhookDefinitionData(ctx context.Context, path path.Path) (cm.WebhookDefinitionData, diag.Diagnostics) {
	return model.toWebhookDefinitionData(ctx, path, types.StringNull())
}

func (model *WebhookModel) toWebhookDefinitionData(ctx context.Context, path path.Path, priorUsername types.String) (cm.WebhookDefinitionData, diag.Diagnostics) {
	var diags diag.Diagnostics

	name, nameDiags := requestRequiredString(model.Name, path.AtName("name"))
	diags.Append(nameDiags...)

	url, urlDiags := requestRequiredString(model.URL, path.AtName("url"))
	diags.Append(urlDiags...)

	active, activeDiags := requestRequiredBool(model.Active, path.AtName("active"))
	diags.Append(activeDiags...)

	httpBasicUsername, httpBasicUsernameDiags := requestNullableString(model.HTTPBasicUsername, path.AtName("http_basic_username"))
	diags.Append(httpBasicUsernameDiags...)

	httpBasicPassword, httpBasicPasswordDiags := requestNullableString(model.HTTPBasicPassword, path.AtName("http_basic_password"))
	diags.Append(httpBasicPasswordDiags...)

	if _, hasUsername := httpBasicUsername.Get(); hasUsername && httpBasicPassword.IsNull() {
		if model.HTTPBasicUsername.Equal(priorUsername) {
			httpBasicUsername.Reset()
			httpBasicPassword.Reset()
		} else {
			diags.AddAttributeError(path.AtName("http_basic_password"), "Missing webhook Basic password",
				"Supply a password when setting a new Basic username.")
		}
	}

	if _, hasPassword := httpBasicPassword.Get(); hasPassword && httpBasicUsername.IsNull() {
		diags.AddAttributeError(path.AtName("http_basic_username"), "Missing webhook Basic username",
			"Supply a username with the password, or omit both credentials to clear Basic authentication.")
	}

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

	headersList, headersListDiags := ToWebhookDefinitionHeaders(path.AtName("headers"), model.Headers)
	diags.Append(headersListDiags...)

	req.Headers = headersList

	transformation, transformationDiags := ToOptNilWebhookDefinitionTransformation(path.AtName("transformation"), model.Transformation)
	diags.Append(transformationDiags...)

	req.Transformation = transformation

	if diags.HasError() {
		return cm.WebhookDefinitionData{}, diags
	}

	return req, diags
}

func (model *WebhookModel) toWebhookDefinitionCreateData(ctx context.Context, root path.Path) (cm.WebhookDefinitionCreateData, diag.Diagnostics) {
	fields, diags := model.ToWebhookDefinitionData(ctx, root)
	if diags.HasError() {
		return cm.WebhookDefinitionCreateData{}, diags
	}

	create := cm.WebhookDefinitionCreateData{
		Name: fields.Name, URL: fields.URL, Topics: fields.Topics, Filters: fields.Filters,
		HttpBasicUsername: fields.HttpBasicUsername, HttpBasicPassword: fields.HttpBasicPassword,
		Transformation: fields.Transformation, Active: fields.Active,
	}
	if fields.Headers != nil {
		create.Headers = make(cm.WebhookDefinitionCreateHeaders, 0, len(fields.Headers))
	}

	for _, header := range fields.Headers {
		value, ok := header.Value.Get()
		if !ok {
			diags.AddAttributeError(root.AtName("headers").AtMapKey(header.Key).AtName("value"),
				"Missing initial webhook header value", "Supply an initial value for each header when creating a webhook.")

			continue
		}

		create.Headers = append(create.Headers, cm.WebhookDefinitionCreateHeader{Key: header.Key, Secret: header.Secret, Value: value})
	}

	return create, diags
}
