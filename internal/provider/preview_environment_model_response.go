package provider

import (
	"context"
	"fmt"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewPreviewEnvironmentModelFromResponse(_ context.Context, previewEnvironment cm.PreviewEnvironment) (PreviewEnvironmentModel, diag.Diagnostics) {
	model, diagnostics, _ := newPreviewEnvironmentModelFromResponse(previewEnvironment)

	return model, diagnostics
}

func newPreviewEnvironmentModelFromResponse(previewEnvironment cm.PreviewEnvironment) (PreviewEnvironmentModel, diag.Diagnostics, diag.Diagnostics) {
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

	seenContentTypeIDs := make(map[string]struct{}, len(previewEnvironment.Configurations))
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
		if _, exists := seenContentTypeIDs[contentTypeID]; exists {
			duplicateDiagnostic := diag.NewAttributeWarningDiagnostic(
				configurationPath,
				"Duplicate content preview configuration response",
				fmt.Sprintf("Contentful returned more than one active configuration for content type %q. Terraform retained the first configuration and omitted the duplicate.", contentTypeID),
			)
			diagnostics.Append(duplicateDiagnostic)
			configurationDiagnostics.Append(duplicateDiagnostic)

			continue
		}

		seenContentTypeIDs[contentTypeID] = struct{}{}

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
// response and restores the exact planned representation only after proving
// every request-owned value is equivalent. ownedIdentity contains only the
// identity values owned by the request endpoint; null values remain
// response-owned.
func ReconcilePreviewEnvironmentMutationResponse(
	_ context.Context,
	previewEnvironment cm.PreviewEnvironment,
	plan PreviewEnvironmentModel,
	ownedIdentity PreviewEnvironmentIdentityModel,
) (PreviewEnvironmentModel, diag.Diagnostics, diag.Diagnostics) {
	state, responseDiagnostics, configurationDiagnostics := newPreviewEnvironmentModelFromResponse(previewEnvironment)
	consistencyDiagnostics := diag.Diagnostics{}
	mismatch := previewEnvironmentMutationValueDiffers(
		path.Root("space_id"),
		"Contentful returned a different content preview platform space ID",
		ownedIdentity.SpaceID,
		state.SpaceID,
		&consistencyDiagnostics,
	)

	mismatch = previewEnvironmentMutationValueDiffers(
		path.Root("preview_environment_id"),
		"Contentful returned a different content preview platform ID",
		ownedIdentity.PreviewEnvironmentID,
		state.PreviewEnvironmentID,
		&consistencyDiagnostics,
	) || mismatch

	// Endpoint identity remains the Terraform target even when Contentful's
	// successful mutation representation contradicts it.
	if !ownedIdentity.SpaceID.IsNull() && !ownedIdentity.SpaceID.IsUnknown() {
		state.SpaceID = ownedIdentity.SpaceID
	}

	if !ownedIdentity.PreviewEnvironmentID.IsNull() && !ownedIdentity.PreviewEnvironmentID.IsUnknown() {
		state.PreviewEnvironmentID = ownedIdentity.PreviewEnvironmentID
	}

	if !state.SpaceID.IsNull() && !state.SpaceID.IsUnknown() &&
		!state.PreviewEnvironmentID.IsNull() && !state.PreviewEnvironmentID.IsUnknown() {
		state.ID = NewIDIdentityModelFromMultipartID(
			state.SpaceID.ValueString(),
			state.PreviewEnvironmentID.ValueString(),
		).ID
	}

	if !plan.Name.Equal(state.Name) {
		consistencyDiagnostics.AddAttributeError(
			path.Root("name"),
			"Contentful returned a different content preview platform name",
			"Contentful accepted the request but returned a name that differs from the value Terraform applied. Terraform retained the returned value in state rather than substituting the planned value.",
		)

		mismatch = true
	}

	if !plan.Description.Equal(state.Description) {
		consistencyDiagnostics.AddAttributeError(
			path.Root("description"),
			"Contentful returned a different content preview platform description",
			"Contentful accepted the request but returned a description that differs from the value Terraform applied. Terraform retained the returned value in state rather than substituting the planned value.",
		)

		mismatch = true
	}

	if len(configurationDiagnostics) != 0 {
		consistencyDiagnostics.AddAttributeError(
			path.Root("content_type_configurations"),
			"Provider cannot fully represent content preview configurations",
			"Contentful accepted the request, but the returned content preview configurations contain values this provider cannot fully represent. Terraform retained the representable response values but cannot verify that they match the value Terraform applied.",
		)

		mismatch = true
	} else if previewEnvironmentConfigurationConsistencyDiffers(plan.ContentTypeConfigurations, state.ContentTypeConfigurations, &consistencyDiagnostics) {
		mismatch = true
	}

	if !mismatch {
		state.Name = plan.Name
		state.Description = plan.Description
		state.ContentTypeConfigurations = plan.ContentTypeConfigurations
	}

	return state, responseDiagnostics, consistencyDiagnostics
}

func previewEnvironmentMutationValueDiffers(
	valuePath path.Path,
	summary string,
	planned types.String,
	remote types.String,
	diagnostics *diag.Diagnostics,
) bool {
	if planned.IsNull() || planned.IsUnknown() || planned.Equal(remote) {
		return false
	}

	diagnostics.AddAttributeError(
		valuePath,
		summary,
		"Contentful accepted the request but returned an identity that differs from the requested endpoint. Terraform retained the requested endpoint identity while preserving the other returned values in recovery state.",
	)

	return true
}

func previewEnvironmentConfigurationConsistencyDiffers(
	planned TypedMap[TypedObject[PreviewEnvironmentContentTypeConfigurationValue]],
	remote TypedMap[TypedObject[PreviewEnvironmentContentTypeConfigurationValue]],
	diagnostics *diag.Diagnostics,
) bool {
	if planned.Equal(remote) {
		return false
	}

	mismatch := false
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

			mismatch = true

			continue
		}

		if !plannedConfiguration.Equal(remoteConfiguration) {
			diagnostics.AddAttributeError(
				configurationPath.AtName("url"),
				"Contentful returned a different content preview URL",
				fmt.Sprintf("Contentful accepted the request but returned a different active preview URL for content type %q. Terraform retained the returned configuration map.", contentTypeID),
			)

			mismatch = true
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

		mismatch = true
	}

	return mismatch
}
