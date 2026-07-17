package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PreviewEnvironmentIdentityModel struct {
	SpaceID              types.String `tfsdk:"space_id"`
	PreviewEnvironmentID types.String `tfsdk:"preview_environment_id"`
}

type PreviewEnvironmentContentTypeConfigurationValue struct {
	URL types.String `tfsdk:"url"`
}

type PreviewEnvironmentModel struct {
	IDIdentityModel
	PreviewEnvironmentIdentityModel

	Name                      types.String                                                           `tfsdk:"name"`
	Description               types.String                                                           `tfsdk:"description"`
	ContentTypeConfigurations TypedMap[TypedObject[PreviewEnvironmentContentTypeConfigurationValue]] `tfsdk:"content_type_configurations"`
	Timeouts                  timeouts.Value                                                         `tfsdk:"timeouts"`
}
