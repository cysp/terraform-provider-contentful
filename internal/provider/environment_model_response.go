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

	model := EnvironmentModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, environmentID),
		EnvironmentIdentityModel: EnvironmentIdentityModel{
			SpaceID:       types.StringValue(spaceID),
			EnvironmentID: types.StringValue(environmentID),
		},
		Status: types.StringValue(environment.Sys.Status.Sys.ID),
	}

	model.Name = types.StringValue(environment.Name)

	return model, diags
}
