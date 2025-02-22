//nolint:dupl
package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *WebhookModel) ToUpdateWebhookDefinitionReq(ctx context.Context, path path.Path) (cm.UpdateWebhookDefinitionReq, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	req := cm.UpdateWebhookDefinitionReq{
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
		diags.Append(model.Topics.ElementsAs(ctx, &topics, false)...)

		req.Topics = topics
	}

	if model.Filters.IsNull() || model.Filters.IsUnknown() {
		req.Filters = cm.NewOptNilWebhookDefinitionFilterArrayNull()
	} else {
		path := path.AtName("filters")

		modelFilters := make([]WebhookFilterValue, len(model.Filters.Elements()))
		diags.Append(model.Filters.ElementsAs(ctx, &modelFilters, false)...)

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

	transformation, transformationDiags := ToOptNilUpdateWebhookDefinitionReqTransformation(ctx, path.AtName("transformation"), model.Transformation)
	diags.Append(transformationDiags...)

	req.Transformation = transformation

	return req, diags
}
