package provider

import (
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func appEventSubscriptionIDs(model AppEventSubscriptionModel) (string, string, diag.Diagnostics) {
	organization, diags := requestRequiredString(model.OrganizationID, path.Root("organization_id"))
	if !diags.HasError() && organization == "" {
		diags.AddAttributeError(path.Root("organization_id"), "Invalid app event subscription identity", "Identity components must not be empty.")
	}

	appDefinition, appDiags := requestRequiredString(model.AppDefinitionID, path.Root("app_definition_id"))
	if !appDiags.HasError() && appDefinition == "" {
		appDiags.AddAttributeError(path.Root("app_definition_id"), "Invalid app event subscription identity", "Identity components must not be empty.")
	}

	diags.Append(appDiags...)

	return organization, appDefinition, diags
}

// ToAppEventSubscriptionData builds the complete requested configuration from
// the effective Plan. Null optional values omit wire members; no JSON null or
// empty functions object is used as an invented clearing instruction.
func (model AppEventSubscriptionModel) ToAppEventSubscriptionData() (cm.AppEventSubscriptionData, diag.Diagnostics) {
	topics, diags := knownOptionalStringSetElements(path.Root("topics"), model.Topics)
	if !diags.HasError() && len(topics) == 0 {
		diags.AddAttributeError(path.Root("topics"), "Invalid app event topics", "At least one topic must be supplied.")
	}

	for _, value := range topics {
		if value == "" {
			diags.AddAttributeError(path.Root("topics").AtSetValue(types.StringValue(value)), "Invalid app event topic", "Topics must not be empty strings.")
		}
	}

	slices.Sort(topics)

	target, targetDiags := requestOmittableString(model.TargetURL, path.Root("target_url"))
	diags.Append(targetDiags...)

	if value, ok := target.Get(); ok && !appEventSubscriptionTargetPattern.MatchString(value) {
		diags.AddAttributeError(path.Root("target_url"), "Invalid app event target URL", "The target must be an HTTPS URL.")
	}

	functions := cm.AppEventSubscriptionFunctions{}
	for _, role := range []struct {
		name   string
		value  types.String
		target *cm.OptFunctionLink
	}{
		{"filter_function_id", model.FilterFunctionID, &functions.Filter},
		{"transformation_function_id", model.TransformationFunctionID, &functions.Transformation},
		{"handler_function_id", model.HandlerFunctionID, &functions.Handler},
	} {
		value, valueDiags := requestOmittableString(role.value, path.Root(role.name))
		diags.Append(valueDiags...)

		if id, ok := value.Get(); ok {
			if id == "" {
				diags.AddAttributeError(path.Root(role.name), "Invalid Function ID", "A configured Function ID must not be empty.")
			} else {
				role.target.SetTo(cm.NewFunctionLink(id))
			}
		}
	}

	data := cm.AppEventSubscriptionData{Topics: topics, TargetUrl: target}
	if functions.Filter.IsSet() || functions.Transformation.IsSet() || functions.Handler.IsSet() {
		data.Functions.SetTo(functions)
	}

	return data, diags
}
