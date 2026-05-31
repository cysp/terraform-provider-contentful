package provider

import (
	"context"
	"maps"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (model *WebhookModel) ToWebhookDefinitionData(ctx context.Context, config WebhookModel, path path.Path) (cm.WebhookDefinitionData, diag.Diagnostics) {
	diags := rejectUnknownConfigurationOwnedRequestValue(model.Headers, config.Headers, path.AtName("headers"))

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

	headersList, headersListDiags := ToWebhookDefinitionHeaders(path.AtName("headers"), model.Headers, config.Headers)
	diags.Append(headersListDiags...)

	req.Headers = headersList

	transformation, transformationDiags := ToOptNilWebhookDefinitionDataTransformation(path.AtName("transformation"), model.Transformation)
	diags.Append(transformationDiags...)

	req.Transformation = transformation

	if diags.HasError() {
		return cm.WebhookDefinitionData{}, diags
	}

	return req, diags
}

func WebhookModelWithWriteOnlySecrets(plan, config WebhookModel) (WebhookModel, WriteOnlySecretValues, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	values := WriteOnlySecretValues{}
	model := plan

	if plan.Headers.IsNull() || plan.Headers.IsUnknown() {
		return model, values, diags
	}

	headers := maps.Clone(plan.Headers.Elements())

	configHeaders := map[string]TypedObject[WebhookHeaderValue]{}
	if !config.Headers.IsNull() && !config.Headers.IsUnknown() {
		configHeaders = config.Headers.Elements()
	}

	headerKeys := slices.Sorted(maps.Keys(headers))
	for _, key := range headerKeys {
		header := headers[key].Value()
		configHeader := WebhookHeaderValue{}

		if configured, ok := configHeaders[key]; ok {
			configHeader = configured.Value()
		}

		valuePath := path.Root("headers").AtMapKey(key).AtName("value")
		valueWOPath := path.Root("headers").AtMapKey(key).AtName("value_wo")

		value, usedWriteOnly, valueDiags := resolveStringSecret(
			configHeader.Value,
			configHeader.ValueWO,
			valuePath,
			valueWOPath,
			true,
		)
		diags.Append(valueDiags...)

		header.Value = value
		header.ValueWO = types.StringNull()

		if usedWriteOnly {
			values.Add(valueWOPath, configHeader.ValueWO)
		}

		headers[key] = NewTypedObject(header)
	}

	model.Headers = NewTypedMap(headers)

	return model, values, diags
}
