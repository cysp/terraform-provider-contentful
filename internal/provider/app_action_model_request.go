package provider

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func appActionScope(model AppActionBaseModel) diag.Diagnostics {
	diags := requireDiscoveryID(path.Root("organization_id"), model.OrganizationID)
	diags.Append(requireDiscoveryID(path.Root("app_definition_id"), model.AppDefinitionID)...)

	return diags
}

func (model AppActionModel) ToAppActionData() (cm.AppActionData, diag.Diagnostics) {
	var data cm.AppActionData

	name, diags := requestRequiredString(model.Name, path.Root("name"))
	data.Name = name
	category, more := requestRequiredString(model.Category, path.Root("category"))
	diags.Append(more...)

	data.Category = category
	kind, more := requestRequiredString(model.Type, path.Root("type"))
	diags.Append(more...)

	data.Type = kind
	data.Description, more = requestOmittableString(model.Description, path.Root("description"))
	diags.Append(more...)
	data.URL, more = requestOmittableString(model.URL, path.Root("url"))
	diags.Append(more...)
	function, more := requestOmittableString(model.FunctionID, path.Root("function_id"))
	diags.Append(more...)

	if id, ok := function.Get(); ok {
		if id == "" {
			diags.AddAttributeError(path.Root("function_id"), "Invalid Function ID", "The Function ID must not be empty.")
		} else {
			data.Function.SetTo(cm.NewFunctionLink(id))
		}
	}

	data.Parameters, more = appActionRequestJSON(model.Parameters, path.Root("parameters"), true)
	diags.Append(more...)

	data.ParametersSchema, more = appActionRequestJSON(model.ParametersSchema, path.Root("parameters_schema"), false)
	diags.Append(more...)
	data.ResultSchema, more = appActionRequestJSON(model.ResultSchema, path.Root("result_schema"), false)
	diags.Append(more...)

	if diags.HasError() {
		return data, diags
	}

	diags.Append(validateAppActionTarget(model.Type, model.URL, model.FunctionID)...)

	if data.Category != "Custom" && data.Category != "Entries.v1.0" && data.Category != "Notification.v1.0" {
		diags.AddAttributeError(path.Root("category"), "Unsupported App Action category", "Supported categories are Custom, Entries.v1.0, and Notification.v1.0.")
	}

	if data.Name == "" || data.Category == "" {
		diags.AddError("Invalid App Action", "Name and category must not be empty.")
	}

	diags.Append(model.validateInput()...)

	return data, diags
}

func (model AppActionModel) validateInput() diag.Diagnostics {
	if model.Category.IsNull() || model.Category.IsUnknown() || model.Parameters.IsUnknown() {
		return nil
	}

	if model.Category.ValueString() != "Custom" {
		if !model.Parameters.IsNull() {
			return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(path.Root("parameters"), "Built-in App Action parameters are read-only", "Only Custom actions accept parameters. Check category, parameters, and lifecycle.ignore_changes.")}
		}

		return nil
	}

	if !model.ParametersSchema.IsUnknown() && model.Parameters.IsNull() == model.ParametersSchema.IsNull() {
		return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(path.Root("parameters"), "Invalid Custom App Action input", "For category Custom, configure either parameters or parameters_schema, but not both. Check lifecycle.ignore_changes if the plan preserves an input definition from state.")}
	}

	return nil
}

func appActionRequestJSON(value jsontypes.Normalized, valuePath path.Path, array bool) (jx.Raw, diag.Diagnostics) {
	if value.IsUnknown() {
		return nil, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(valuePath, "Unknown App Action value", "The parameter or schema value must be known before creating or updating the action.")}
	}

	if value.IsNull() {
		return nil, nil
	}

	raw := value.ValueString()
	diags := validateAppActionJSON(raw, valuePath, array)

	return jx.Raw(raw), diags
}

func validateAppActionJSON(raw string, valuePath path.Path, array bool) diag.Diagnostics {
	root := "{"
	shape := "object"

	if array {
		root = "["
		shape = "array"
	}

	if !json.Valid([]byte(raw)) || !strings.HasPrefix(strings.TrimSpace(raw), root) {
		return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(valuePath, "Invalid App Action JSON", "Provide a JSON "+shape+". To omit this attribute, leave it out of the configuration instead of using the JSON string \"null\".")}
	}

	return nil
}

type appActionJSONValidator struct{ array bool }

func (v appActionJSONValidator) Description(context.Context) string {
	return "Requires a JSON object, or a legacy parameter array for parameters."
}

func (v appActionJSONValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v appActionJSONValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	resp.Diagnostics.Append(validateAppActionJSON(req.ConfigValue.ValueString(), req.Path, v.array)...)
}

func validateAppActionTarget(kind, targetURL, functionID types.String) diag.Diagnostics {
	var diags diag.Diagnostics

	if !targetURL.IsNull() && !targetURL.IsUnknown() {
		target, err := url.Parse(targetURL.ValueString())
		if err != nil || target.Scheme != "https" || target.Host == "" {
			diags.AddAttributeError(path.Root("url"), "Invalid App Action URL", "The endpoint must be an HTTPS URL with a host.")
		}
	}

	if kind.IsNull() || kind.IsUnknown() || targetURL.IsUnknown() || functionID.IsUnknown() {
		return diags
	}

	switch kind.ValueString() {
	case "endpoint":
		if targetURL.IsNull() || !functionID.IsNull() {
			diags.AddAttributeError(path.Root("url"), "Invalid App Action configuration", "For type endpoint, set url and omit function_id.")
		}
	case "function-invocation":
		if functionID.IsNull() || !targetURL.IsNull() {
			diags.AddAttributeError(path.Root("function_id"), "Invalid App Action configuration", "For type function-invocation, set function_id and omit url.")
		}
	default:
		diags.AddAttributeError(path.Root("type"), "Invalid App Action configuration", "Set type to endpoint or function-invocation.")
	}

	return diags
}
