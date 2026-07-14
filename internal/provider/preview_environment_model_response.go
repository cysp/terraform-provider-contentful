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
	spaceID := previewEnvironment.Sys.Space.Sys.ID
	previewEnvironmentID := previewEnvironment.Sys.ID
	diagnostics := diag.Diagnostics{}

	configurations := make(map[string]TypedObject[PreviewEnvironmentContentTypeConfigurationValue], len(previewEnvironment.Configurations))

	seenContentTypeIDs := make(map[string]struct{}, len(previewEnvironment.Configurations))
	for index, configuration := range previewEnvironment.Configurations {
		contentTypeID, identityDiagnostics := previewEnvironmentContentTypeIDFromResponse(index, configuration)
		diagnostics.Append(identityDiagnostics...)

		if identityDiagnostics.HasError() {
			continue
		}

		configurationPath := path.Root("content_type_configurations").AtMapKey(contentTypeID)
		if _, exists := seenContentTypeIDs[contentTypeID]; exists {
			diagnostics.AddAttributeError(
				configurationPath,
				"Duplicate content preview configuration response",
				fmt.Sprintf("Contentful returned more than one configuration for content type %q.", contentTypeID),
			)

			continue
		}

		seenContentTypeIDs[contentTypeID] = struct{}{}

		if !configuration.Enabled {
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
	}, diagnostics
}

func previewEnvironmentContentTypeIDFromResponse(
	index int,
	configuration cm.PreviewEnvironmentConfiguration,
) (string, diag.Diagnostics) {
	diagnostics := diag.Diagnostics{}
	configurationsPath := path.Root("content_type_configurations")

	entityType := configuration.EntityType.Or("ContentType")
	if entityType != "ContentType" {
		diagnostics.AddAttributeError(
			configurationsPath,
			"Unsupported content preview configuration response",
			fmt.Sprintf("Configuration %d has unsupported entity type %q; only ContentType is supported.", index, entityType),
		)

		return "", diagnostics
	}

	entityID := configuration.EntityId.Or("")

	contentTypeID := configuration.ContentType.Or("")
	if entityID != "" && contentTypeID != "" && entityID != contentTypeID {
		diagnostics.AddAttributeError(
			configurationsPath,
			"Conflicting content preview configuration response",
			fmt.Sprintf("Configuration %d has entityId %q but contentType %q.", index, entityID, contentTypeID),
		)

		return "", diagnostics
	}

	if entityID == "" {
		entityID = contentTypeID
	}

	if entityID == "" {
		diagnostics.AddAttributeError(
			configurationsPath,
			"Invalid content preview configuration response",
			fmt.Sprintf("Configuration %d has neither entityId nor contentType.", index),
		)
	}

	return entityID, diagnostics
}
