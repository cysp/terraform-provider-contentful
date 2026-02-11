package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEnvironmentStatusReadyModelFromResponse(_ context.Context, environment cm.Environment) (EnvironmentStatusReadyModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	spaceID := environment.Sys.Space.Sys.ID
	environmentID := environment.Sys.ID

	model := EnvironmentStatusReadyModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, environmentID),
		EnvironmentIdentityModel: EnvironmentIdentityModel{
			SpaceID:       types.StringValue(spaceID),
			EnvironmentID: types.StringValue(environmentID),
		},
		Status: types.StringValue(environment.Sys.Status.Sys.ID),
	}

	return model, diags
}
