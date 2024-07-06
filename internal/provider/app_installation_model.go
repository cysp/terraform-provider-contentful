package provider

import (
	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/resource_app_installation"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func ReadAppInstallationModel(model *resource_app_installation.AppInstallationModel, appInstallation contentfulManagement.AppInstallation) {
	// SpaceId, EnvironmentId and AppDefinitionId are all already known
	if appInstallation.Parameters.Set {
		encoder := jx.Encoder{}
		appInstallation.Parameters.Encode(&encoder)
		model.Parameters = jsontypes.NewNormalizedValue(encoder.String())
	} else {
		model.Parameters = jsontypes.NewNormalizedNull()
	}
}

func CreatePutAppInstallationRequestBody(req *contentfulManagement.PutAppInstallationReq, model resource_app_installation.AppInstallationModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	switch {
	case model.Parameters.IsUnknown():
		diags.AddAttributeWarning(path.Root("parameters"), "Failed to update app installation parameters", "Parameters are unknown")
		req.Parameters.Reset()
	case model.Parameters.IsNull():
		req.Parameters.Reset()
	default:
		appInstallationParametersValue := contentfulManagement.PutAppInstallationReqParameters{}
		diags.Append(model.Parameters.Unmarshal(&appInstallationParametersValue)...)
		req.Parameters.SetTo(appInstallationParametersValue)
	}

	return diags
}
