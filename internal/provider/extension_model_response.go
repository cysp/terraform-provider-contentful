package provider

import (
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (model *ExtensionResourceModel) ReadFromResponse(extension *cm.Extension) diag.Diagnostics {
	diags := diag.Diagnostics{}

	spaceID := extension.Sys.Space.Sys.ID
	environmentID := extension.Sys.Environment.Sys.ID
	extensionID := extension.Sys.ID

	model.ID = types.StringValue(strings.Join([]string{spaceID, environmentID, extensionID}, "/"))
	model.SpaceID = types.StringValue(spaceID)
	model.EnvironmentID = types.StringValue(environmentID)
	model.ExtensionID = types.StringValue(extensionID)

	if extension.Extension.Parameters != nil {
		constraint, err := util.JxNormalizeOpaqueBytes(extension.Extension.Parameters, util.JxEncodeOpaqueOptions{EscapeStrings: true})
		if err != nil {
			diags.AddAttributeError(path.Root("parameters"), "Failed to read parameters", err.Error())
		}

		model.Parameters = jsontypes.NewNormalizedValue(string(constraint))
	} else {
		model.Parameters = jsontypes.NewNormalizedNull()
	}

	return diags
}
