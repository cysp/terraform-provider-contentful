package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewAppSigningSecretResourceModelFromResponse(res cm.AppSigningSecret, organizationID, appDefinitionID string, value types.String) (AppSigningSecretModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if res.Sys.Organization.Sys.ID != organizationID || organizationID == "" {
		diags.AddAttributeError(path.Root("organization_id"), "Unexpected app signing secret identity", "Contentful did not return the requested organization identity. No new state was published.")
	}

	if res.Sys.AppDefinition.Sys.ID != appDefinitionID || appDefinitionID == "" {
		diags.AddAttributeError(path.Root("app_definition_id"), "Unexpected app signing secret identity", "Contentful did not return the requested App Definition identity. No new state was published.")
	}

	if value.IsUnknown() {
		diags.AddAttributeError(path.Root("value"), "Unexpected unknown signing secret value", "The retained signing secret value must be known or null.")
	}

	if diags.HasError() {
		return AppSigningSecretModel{}, diags
	}

	model := AppSigningSecretModel{
		Value:           value,
		ValueWO:         types.StringNull(),
		IDIdentityModel: NewIDIdentityModelFromMultipartID(organizationID, appDefinitionID),
		AppSigningSecretIdentityModel: AppSigningSecretIdentityModel{
			OrganizationID:  types.StringValue(organizationID),
			AppDefinitionID: types.StringValue(appDefinitionID),
		},
	}

	model.CreatedAt = timetypes.NewRFC3339TimePointerValue(res.Sys.CreatedAt.ValueTimePointer())
	model.UpdatedAt = timetypes.NewRFC3339TimePointerValue(res.Sys.UpdatedAt.ValueTimePointer())

	return model, diags
}
