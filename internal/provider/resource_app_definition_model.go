package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
)

type AppDefinitionResourceModel struct {
	AppDefinitionBaseModel

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
