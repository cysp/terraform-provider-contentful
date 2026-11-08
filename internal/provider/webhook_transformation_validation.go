package provider

import (
	"context"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type webhookTransformationStringValidator struct {
	values []string
}

func webhookTransformationMethodValidator() webhookTransformationStringValidator {
	return webhookTransformationStringValidator{values: []string{"POST", "GET", "PUT", "PATCH", "DELETE"}}
}

func webhookTransformationContentTypeValidator() webhookTransformationStringValidator {
	return webhookTransformationStringValidator{values: []string{
		"application/vnd.contentful.management.v1+json",
		"application/vnd.contentful.management.v1+json; charset=utf-8",
		"application/json", "application/json; charset=utf-8",
		"application/x-www-form-urlencoded", "application/x-www-form-urlencoded; charset=utf-8",
	}}
}

func (v webhookTransformationStringValidator) Description(context.Context) string {
	return v.description()
}

func (v webhookTransformationStringValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v webhookTransformationStringValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	resp.Diagnostics.Append(v.validate(req.ConfigValue, req.Path)...)
}

func (v webhookTransformationStringValidator) description() string {
	return "When configured, the value must be one of: " + strings.Join(v.values, ", ") + ". Omit it to use Contentful's default."
}

func (v webhookTransformationStringValidator) validate(value types.String, valuePath path.Path) diag.Diagnostics {
	if value.IsNull() || value.IsUnknown() || slices.Contains(v.values, value.ValueString()) {
		return nil
	}

	return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(valuePath, "Invalid webhook transformation value", v.description())}
}
