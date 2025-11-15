package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *TeamModel) ToTeamData(_ context.Context, _ path.Path) (cm.TeamData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fields := cm.TeamData{
		Name: model.Name.ValueString(),
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		fields.Description = cm.NewNilString(model.Description.ValueString())
	} else {
		fields.Description = cm.NewNilStringNull()
	}

	return fields, diags
}
