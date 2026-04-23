package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
)

type AppDefinitionDataSourceModel struct {
	AppDefinitionBaseModel

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
