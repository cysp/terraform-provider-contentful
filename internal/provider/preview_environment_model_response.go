package provider

import (
	"context"
	"fmt"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewPreviewEnvironmentResourceModelFromResponse(previewEnvironment cm.PreviewEnvironment) (PreviewEnvironmentModel, diag.Diagnostics) {
	model, diagnostics, _ := newPreviewEnvironmentResourceModelFromResponse(previewEnvironment)

	return model, diagnostics
}

func newPreviewEnvironmentResourceModelFromResponse(previewEnvironment cm.PreviewEnvironment) (PreviewEnvironmentModel, diag.Diagnostics, diag.Diagnostics) {
	spaceID := previewEnvironment.Sys.Space.Sys.ID
	previewEnvironmentID := previewEnvironment.Sys.ID
	diagnostics := diag.Diagnostics{}
	configurationDiagnostics := diag.Diagnostics{}

	if spaceID == "" {
		diagnostics.AddAttributeError(
			path.Root("space_id"),
			"Missing content preview platform space ID",
			"Contentful returned a content preview platform without the space identity required to publish Terraform state.",
		)
	}

	if previewEnvironmentID == "" {
		diagnostics.AddAttributeError(
			path.Root("preview_environment_id"),
			"Missing content preview platform ID",
			"Contentful returned a content preview platform without the resource identity required to publish Terraform state.",
		)
	}

	configurations := make(map[string]TypedObject[PreviewEnvironmentContentTypeConfigurationValue], len(previewEnvironment.Configurations))
	for index, configuration := range previewEnvironment.Configurations {
		if !configuration.Enabled {
			continue
		}

		contentTypeID, identityDiagnostics := previewEnvironmentContentTypeIDFromResponse(index, configuration)
		diagnostics.Append(identityDiagnostics...)
		configurationDiagnostics.Append(identityDiagnostics...)

		if len(identityDiagnostics) != 0 {
			continue
		}

		configurationPath := path.Root("content_type_configurations").AtMapKey(contentTypeID)
		if _, exists := configurations[contentTypeID]; exists {
			duplicateDiagnostic := diag.NewAttributeWarningDiagnostic(
				configurationPath,
				"Duplicate content preview configuration response",
				fmt.Sprintf("Contentful returned more than one active configuration for content type %q. Terraform retained the first configuration and omitted the duplicate.", contentTypeID),
			)
			diagnostics.Append(duplicateDiagnostic)
			configurationDiagnostics.Append(duplicateDiagnostic)

			continue
		}

		configurations[contentTypeID] = NewTypedObject(PreviewEnvironmentContentTypeConfigurationValue{
			URL: types.StringValue(configuration.URL),
		})
	}

	return PreviewEnvironmentModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, previewEnvironmentID),
		PreviewEnvironmentIdentityModel: PreviewEnvironmentIdentityModel{
			SpaceID:              types.StringValue(spaceID),
			PreviewEnvironmentID: types.StringValue(previewEnvironmentID),
		},
		Name:                      types.StringValue(previewEnvironment.Name),
		Description:               types.StringValue(previewEnvironment.Description),
		ContentTypeConfigurations: NewTypedMap(configurations),
	}, diagnostics, configurationDiagnostics
}

func previewEnvironmentContentTypeIDFromResponse(
	index int,
	configuration cm.PreviewEnvironmentConfiguration,
) (string, diag.Diagnostics) {
	diagnostics := diag.Diagnostics{}
	configurationsPath := path.Root("content_type_configurations")

	entityID := configuration.EntityId.Or("")
	contentTypeID := configuration.ContentType.Or("")

	configurationPath := configurationsPath
	if entityID != "" {
		configurationPath = configurationPath.AtMapKey(entityID)
	} else if contentTypeID != "" {
		configurationPath = configurationPath.AtMapKey(contentTypeID)
	}

	entityType := configuration.EntityType.Or("ContentType")
	if entityType != "ContentType" {
		diagnostics.AddAttributeWarning(
			configurationPath,
			"Unsupported content preview configuration response",
			fmt.Sprintf("Configuration %d has unsupported entity type %q; Terraform omitted this configuration because only ContentType configurations are representable.", index, entityType),
		)

		return "", diagnostics
	}

	if entityID != "" && contentTypeID != "" && entityID != contentTypeID {
		diagnostics.AddAttributeWarning(
			configurationPath,
			"Conflicting content preview configuration response",
			fmt.Sprintf("Configuration %d has entityId %q but contentType %q. Terraform omitted this configuration because its identity is ambiguous.", index, entityID, contentTypeID),
		)

		return "", diagnostics
	}

	if entityID == "" {
		entityID = contentTypeID
	}

	if entityID == "" {
		diagnostics.AddAttributeWarning(
			configurationsPath,
			"Invalid content preview configuration response",
			fmt.Sprintf("Configuration %d has neither entityId nor contentType. Terraform omitted this configuration because it cannot be represented as a map element.", index),
		)
	}

	return entityID, diagnostics
}

// ReconcilePreviewEnvironmentMutationResponse projects the complete mutation
// response and checks every known effective Plan value. Exact equality is
// sufficient for the active-configuration map. ownedIdentity contains only the
// identity values owned by the request endpoint; null values remain
// response-owned.
func ReconcilePreviewEnvironmentMutationResponse(
	_ context.Context,
	previewEnvironment cm.PreviewEnvironment,
	plan PreviewEnvironmentModel,
	ownedIdentity PreviewEnvironmentIdentityModel,
) (PreviewEnvironmentModel, diag.Diagnostics, diag.Diagnostics) {
	state, responseDiagnostics, configurationDiagnostics := newPreviewEnvironmentResourceModelFromResponse(previewEnvironment)
	reconciler := mutationResponseReconciler{resourceName: "content preview platform"}
	reconciler.pinIdentity(path.Root("space_id"), ownedIdentity.SpaceID, state.SpaceID, &state.SpaceID)
	reconciler.pinIdentity(path.Root("preview_environment_id"), ownedIdentity.PreviewEnvironmentID, state.PreviewEnvironmentID, &state.PreviewEnvironmentID)

	state.IDIdentityModel = NewIDIdentityModelFromMultipartID(state.SpaceID.ValueString(), state.PreviewEnvironmentID.ValueString())

	if ownedIdentity.PreviewEnvironmentID.IsNull() || ownedIdentity.PreviewEnvironmentID.IsUnknown() {
		if !plan.PreviewEnvironmentID.IsNull() {
			reconciler.compareExact(path.Root("preview_environment_id"), "Contentful returned a different content preview platform ID", plan.PreviewEnvironmentID, state.PreviewEnvironmentID)
		}
	}

	if !plan.ID.IsNull() && !plan.ID.IsUnknown() && !plan.ID.Equal(state.ID) {
		reconciler.diagnostics.AddAttributeError(path.Root("id"), "Content preview platform identity is inconsistent with its endpoint", "The planned legacy ID differs from the content preview platform endpoint identity. Terraform retained the endpoint identity as the resource target and the remaining returned values as recovery state. Review or re-import the content preview platform before applying again.")
	}

	reconciler.compareExact(path.Root("name"), "Contentful returned a different content preview platform name", plan.Name, state.Name)
	reconciler.compareExact(path.Root("description"), "Contentful returned a different content preview platform description", plan.Description, state.Description)

	if !plan.ContentTypeConfigurations.IsUnknown() {
		if len(configurationDiagnostics) != 0 {
			reconciler.diagnostics.AddAttributeError(
				path.Root("content_type_configurations"),
				"Provider cannot fully represent content preview configurations",
				"Contentful accepted the request, but the returned content preview configurations contain values this provider cannot fully represent. Terraform retained the representable response values but cannot verify that they match the value Terraform applied. Review the content preview platform in Contentful before applying again.",
			)
		} else {
			checkPreviewEnvironmentConfigurationConsistency(plan.ContentTypeConfigurations, state.ContentTypeConfigurations, &reconciler.diagnostics)
		}
	}

	return state, responseDiagnostics, reconciler.diagnostics
}

func checkPreviewEnvironmentConfigurationConsistency(
	planned TypedMap[TypedObject[PreviewEnvironmentContentTypeConfigurationValue]],
	remote TypedMap[TypedObject[PreviewEnvironmentContentTypeConfigurationValue]],
	diagnostics *diag.Diagnostics,
) {
	if planned.Equal(remote) {
		return
	}

	if planned.IsNull() {
		diagnostics.AddAttributeError(path.Root("content_type_configurations"), "Contentful returned unexpected content preview configurations", "Contentful accepted the request but returned configurations for a null planned value. Terraform retained the returned configuration map.")

		return
	}

	configurationsPath := path.Root("content_type_configurations")

	for contentTypeID, plannedConfiguration := range planned.Elements() {
		configurationPath := configurationsPath.AtMapKey(contentTypeID)
		remoteConfiguration, exists := remote.Elements()[contentTypeID]

		if !exists {
			diagnostics.AddAttributeError(
				configurationPath,
				"Contentful omitted a planned content preview configuration",
				fmt.Sprintf("Contentful accepted the request but did not return the planned active configuration for content type %q. Terraform retained the returned configuration map.", contentTypeID),
			)

			continue
		}

		if !plannedConfiguration.Equal(remoteConfiguration) {
			diagnostics.AddAttributeError(
				configurationPath.AtName("url"),
				"Contentful returned a different content preview URL",
				fmt.Sprintf("Contentful accepted the request but returned a different active preview URL for content type %q. Terraform retained the returned configuration map.", contentTypeID),
			)
		}
	}

	for contentTypeID := range remote.Elements() {
		if _, exists := planned.Elements()[contentTypeID]; exists {
			continue
		}

		diagnostics.AddAttributeError(
			configurationsPath.AtMapKey(contentTypeID),
			"Contentful returned an unexpected content preview configuration",
			fmt.Sprintf("Contentful accepted the request but returned an active configuration for unplanned content type %q. Terraform retained the returned configuration map.", contentTypeID),
		)
	}
}
