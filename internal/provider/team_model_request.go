package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *TeamModel) ToTeamData(_ context.Context, modelPath path.Path) (cm.TeamData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fields := cm.TeamData{
		Name: model.Name.ValueString(),
	}

	switch {
	case model.Description.IsUnknown():
		diags.AddAttributeError(modelPath.AtName("description"), "Unexpected unknown team description", "The optional team description must be known before it can be sent to Contentful.")
	case !model.Description.IsNull():
		fields.Description = cm.NewNilString(model.Description.ValueString())
	default:
		fields.Description = cm.NewNilStringNull()
	}

	return fields, diags
}
