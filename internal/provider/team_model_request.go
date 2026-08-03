package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *TeamModel) ToTeamData(valuePath path.Path) (cm.TeamData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	name, nameDiags := requestRequiredString(model.Name, valuePath.AtName("name"))
	diags.Append(nameDiags...)

	description := cm.NilString{}

	switch {
	case model.Description.IsUnknown():
		diags.AddAttributeError(
			valuePath.AtName("description"),
			"Unexpected unknown string",
			"The string value must be known before it can be sent to Contentful.",
		)
	case model.Description.IsNull():
		description = cm.NewNilStringNull()
	default:
		description = cm.NewNilString(model.Description.ValueString())
	}

	if diags.HasError() {
		return cm.TeamData{}, diags
	}

	return cm.TeamData{
		Name:        name,
		Description: description,
	}, diags
}
