package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewAppSigningSecretResourceModelFromResponse(res cm.AppSigningSecret) AppSigningSecretModel {
	organizationID := res.Sys.Organization.Sys.ID
	appDefinitionID := res.Sys.AppDefinition.Sys.ID

	model := AppSigningSecretModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(organizationID, appDefinitionID),
		OrganizationID:  types.StringValue(organizationID),
		AppDefinitionID: types.StringValue(appDefinitionID),
	}

	return model
}

// A response from another singleton does not acknowledge the secret sent to the
// requested endpoint. Callers retain prior state on failure; Create has none.
func appSigningSecretMutationResponse(response cm.AppSigningSecret, plan AppSigningSecretModel) (AppSigningSecretModel, diag.Diagnostics) {
	model := NewAppSigningSecretResourceModelFromResponse(response)
	reconciler := mutationResponseReconciler{resourceName: "app signing secret"}
	reconciler.pinIdentity(path.Root("organization_id"), plan.OrganizationID, model.OrganizationID, &model.OrganizationID)
	reconciler.pinIdentity(path.Root("app_definition_id"), plan.AppDefinitionID, model.AppDefinitionID, &model.AppDefinitionID)

	if reconciler.identityMismatch {
		return model, diag.Diagnostics{diag.NewErrorDiagnostic("Unconfirmed app signing secret mutation", "Contentful returned an identity that differs from the requested App Signing Secret endpoint. Terraform did not acknowledge the new secret value. Any prior secret state was retained. Check the secret in Contentful before applying again.")}
	}

	model.Value = plan.Value

	return model, nil
}
