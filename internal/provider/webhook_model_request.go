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
	model.HeaderValuesWO = NewTypedMapNull[types.String]()

	if plan.Headers.IsNull() {
		if !config.HeaderValuesWO.IsNull() && !config.HeaderValuesWO.IsUnknown() {
			for key := range config.HeaderValuesWO.Elements() {
				diags.AddAttributeError(
					path.Root("header_values_wo").AtMapKey(key),
					"Unknown webhook header",
					"Write-only header values must use the same keys configured in headers.",
				)
			}
		}

		return model, values, diags
	}

	if plan.Headers.IsUnknown() {
		return model, values, diags
	}

	headers := maps.Clone(plan.Headers.Elements())

	configHeaders := map[string]TypedObject[WebhookHeaderValue]{}

	configHeadersUnknown := config.Headers.IsUnknown()
	if !config.Headers.IsNull() && !config.Headers.IsUnknown() {
		configHeaders = config.Headers.Elements()
	}

	configHeaderValuesWO := map[string]types.String{}

	configHeaderValuesWOUnknown := config.HeaderValuesWO.IsUnknown()
	if !config.HeaderValuesWO.IsNull() && !config.HeaderValuesWO.IsUnknown() {
		configHeaderValuesWO = config.HeaderValuesWO.Elements()
	}

	for key := range configHeaderValuesWO {
		if _, ok := headers[key]; !ok {
			diags.AddAttributeError(
				path.Root("header_values_wo").AtMapKey(key),
				"Unknown webhook header",
				"Write-only header values must use the same keys configured in headers.",
			)
		}
	}

	headerKeys := slices.Sorted(maps.Keys(headers))
	for _, key := range headerKeys {
		header, writeOnlySecret, hasWriteOnlySecret, headerDiags := webhookHeaderWithWriteOnlySecret(
			key,
			headers[key].Value(),
			configHeaders,
			configHeadersUnknown,
			configHeaderValuesWO,
			configHeaderValuesWOUnknown,
		)
		diags.Append(headerDiags...)

		if hasWriteOnlySecret {
			values = append(values, writeOnlySecret)
		}

		headers[key] = NewTypedObject(header)
	}

	model.Headers = NewTypedMap(headers)

	return model, values, diags
}

func webhookHeaderWithWriteOnlySecret(
	key string,
	header WebhookHeaderValue,
	configHeaders map[string]TypedObject[WebhookHeaderValue],
	configHeadersUnknown bool,
	configHeaderValuesWO map[string]types.String,
	configHeaderValuesWOUnknown bool,
) (WebhookHeaderValue, WriteOnlySecretValue, bool, diag.Diagnostics) {
	configuredHeader, headerConfigured := configHeaders[key]
	configHeaderValueWO, writeOnlyConfigured := configHeaderValuesWO[key]

	if !configHeadersUnknown && !configHeaderValuesWOUnknown && !headerConfigured && !writeOnlyConfigured {
		return header, WriteOnlySecretValue{}, false, nil
	}

	configHeader := header
	if configHeadersUnknown {
		configHeader.Value = types.StringUnknown()
	} else if headerConfigured {
		configHeader = configuredHeader.Value()
	}

	if configHeaderValuesWOUnknown {
		configHeaderValueWO = types.StringUnknown()
	}

	valuePath := path.Root("headers").AtMapKey(key).AtName("value")
	valueWOPath := path.Root("header_values_wo").AtMapKey(key)
	value, usedWriteOnly, diags := resolveStringSecret(
		configHeader.Value,
		configHeaderValueWO,
		valuePath,
		valueWOPath,
	)
	header.Value = value

	if !usedWriteOnly {
		return header, WriteOnlySecretValue{}, false, diags
	}

	if header.Secret.IsNull() || (!header.Secret.IsUnknown() && !header.Secret.ValueBool()) {
		diags.AddAttributeError(
			valueWOPath,
			"Invalid write-only webhook header",
			"Write-only header values can only be configured for headers with secret set to true.",
		)

		return header, WriteOnlySecretValue{}, false, diags
	}

	return header, WriteOnlySecretValue{Path: valueWOPath, Value: configHeaderValueWO}, true, diags
}
