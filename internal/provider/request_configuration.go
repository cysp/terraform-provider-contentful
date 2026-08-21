package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// rejectUnknownConfigurationOwnedRequestValue rejects an unresolved planned
// value when the corresponding attribute is present in resource configuration.
// Resource Create and Update configuration is known by the time the lifecycle
// method runs; configuration is used only to distinguish ownership/presence.
// The request value itself always comes from planned.
func rejectUnknownConfigurationOwnedRequestValue[T attr.Value](
	planned T,
	configured T,
	valuePath path.Path,
) diag.Diagnostics {
	if !planned.IsUnknown() {
		return nil
	}

	if configured.IsNull() {
		return nil
	}

	return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(
		valuePath,
		"Unknown planned value for configuration-owned attribute",
		"This attribute is present in configuration, but its planned value is still unknown. The planned value must be known before it can be sent to Contentful; the provider will not substitute the configuration value.",
	)}
}
