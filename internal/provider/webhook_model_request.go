package provider

import (
	"context"
	"maps"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (model *WebhookModel) ToWebhookDefinitionData(ctx context.Context, path path.Path) (cm.WebhookDefinitionData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	req := cm.WebhookDefinitionData{
		Name:              model.Name.ValueString(),
		URL:               model.URL.ValueString(),
		Active:            util.BoolValueToOptBool(model.Active),
		HttpBasicUsername: cm.NewOptNilPointerString(model.HTTPBasicUsername.ValueStringPointer()),
		HttpBasicPassword: cm.NewOptNilPointerString(model.HTTPBasicPassword.ValueStringPointer()),
	}

	if model.Topics.IsNull() || model.Topics.IsUnknown() {
		req.Topics = nil
	} else {
		topics := make([]string, len(model.Topics.Elements()))
		diags.Append(tfsdk.ValueAs(ctx, model.Topics, &topics)...)

		req.Topics = topics
	}

	if model.Filters.IsNull() || model.Filters.IsUnknown() {
		req.Filters = cm.NewOptNilWebhookDefinitionFilterArrayNull()
	} else {
		path := path.AtName("filters")

		modelFilters := make([]WebhookFilterValue, len(model.Filters.Elements()))
		diags.Append(tfsdk.ValueAs(ctx, model.Filters, &modelFilters)...)

		filters := make([]cm.WebhookDefinitionFilter, len(modelFilters))

		for index, modelFilter := range modelFilters {
			path := path.AtListIndex(index)

			filter, filterDiags := ToWebhookDefinitionFilter(ctx, path, modelFilter)
			diags.Append(filterDiags...)

			filters[index] = filter
		}

		req.Filters = cm.NewOptNilWebhookDefinitionFilterArray(filters)
	}

	headersList, headersListDiags := ToWebhookDefinitionHeaders(ctx, path.AtName("headers"), model.Headers)
	diags.Append(headersListDiags...)

	req.Headers = headersList

	transformation, transformationDiags := ToOptNilWebhookDefinitionDataTransformation(ctx, path.AtName("transformation"), model.Transformation)
	diags.Append(transformationDiags...)

	req.Transformation = transformation

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
