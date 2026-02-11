package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEnvironmentResourceModelFromResponse(_ context.Context, environment cm.Environment) (EnvironmentModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	spaceID := environment.Sys.Space.Sys.ID
	environmentID := environment.Sys.ID

	status := types.StringNull()
	if environmentStatus, ok := environment.Sys.Status.Get(); ok {
		status = types.StringValue(environmentStatus.Sys.ID)
	}

	model := EnvironmentModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, environmentID),
		EnvironmentIdentityModel: EnvironmentIdentityModel{
			SpaceID:       types.StringValue(spaceID),
			EnvironmentID: types.StringValue(environmentID),
		},
		Status: status,
	}

	model.Name = types.StringValue(environment.Name)

	return model, diags
}
