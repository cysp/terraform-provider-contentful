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
		URL:               model.Url.ValueString(),
		Active:            util.BoolValueToOptBool(model.Active),
		HttpBasicUsername: cm.NewOptNilPointerString(model.HttpBasicUsername.ValueStringPointer()),
		HttpBasicPassword: cm.NewOptNilPointerString(model.HttpBasicPassword.ValueStringPointer()),
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

func ToOptNilUpdateWebhookDefinitionReqTransformation(_ context.Context, _ path.Path, value WebhookTransformationValue) (cm.OptNilUpdateWebhookDefinitionReqTransformation, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	optNilTransformation := cm.OptNilUpdateWebhookDefinitionReqTransformation{}

	switch {
	case value.IsUnknown():
		//
	case value.IsNull():
		optNilTransformation.SetToNull()
	default:
		transformation := cm.UpdateWebhookDefinitionReqTransformation{}

		transformation.Method = cm.NewOptPointerString(value.Method.ValueStringPointer())
		transformation.ContentType = cm.NewOptPointerString(value.ContentType.ValueStringPointer())
		transformation.IncludeContentLength = cm.NewOptPointerBool(value.IncludeContentLength.ValueBoolPointer())

		//nolint:revive
		if value.Body.IsUnknown() {
		} else {
			bodyStringPointer := value.Body.ValueStringPointer()
			if bodyStringPointer != nil {
				transformation.Body = []byte(*bodyStringPointer)
			} else {
				transformation.Body = nil
			}
		}

		optNilTransformation.SetTo(transformation)
	}

	return optNilTransformation, diags
}
