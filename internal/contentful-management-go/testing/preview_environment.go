package cmtesting

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewPreviewEnvironmentFromData(spaceID, previewEnvironmentID string, data cm.PreviewEnvironmentData) cm.PreviewEnvironment {
	return cm.PreviewEnvironment{
		Sys:            cm.NewPreviewEnvironmentSys(spaceID, previewEnvironmentID),
		Name:           data.Name,
		Description:    data.Description,
		Configurations: normalizePreviewEnvironmentConfigurations(data.Configurations),
	}
}

func UpdatePreviewEnvironmentFromData(previewEnvironment *cm.PreviewEnvironment, data cm.PreviewEnvironmentData) {
	if previewEnvironment.Name != data.Name || previewEnvironment.Description != data.Description {
		previewEnvironment.Sys.Version++
	}

	previewEnvironment.Name = data.Name
	previewEnvironment.Description = data.Description

	configurationIndexes := make(map[previewEnvironmentConfigurationIdentity]int, len(previewEnvironment.Configurations))
	for index, configuration := range previewEnvironment.Configurations {
		configurationIndexes[previewEnvironmentConfigurationIdentity{
			EntityType: configuration.EntityType.Or("ContentType"),
			EntityID:   configuration.EntityId.Or(configuration.ContentType.Or("")),
		}] = index
	}

	for _, configuration := range normalizePreviewEnvironmentConfigurations(data.Configurations) {
		configurationIdentity := previewEnvironmentConfigurationIdentity{
			EntityType: configuration.EntityType.Or("ContentType"),
			EntityID:   configuration.EntityId.Or(configuration.ContentType.Or("")),
		}
		if index, ok := configurationIndexes[configurationIdentity]; ok {
			previewEnvironment.Configurations[index] = configuration

			continue
		}

		configurationIndexes[configurationIdentity] = len(previewEnvironment.Configurations)
		previewEnvironment.Configurations = append(previewEnvironment.Configurations, configuration)
	}
}

func normalizePreviewEnvironmentConfigurations(configurations []cm.PreviewEnvironmentConfigurationData) []cm.PreviewEnvironmentConfiguration {
	result := make([]cm.PreviewEnvironmentConfiguration, 0, len(configurations))
	for _, configuration := range configurations {
		result = append(result, cm.PreviewEnvironmentConfiguration{
			URL:         configuration.URL,
			EntityType:  cm.NewOptString(configuration.EntityType),
			EntityId:    cm.NewOptString(configuration.EntityId),
			Enabled:     configuration.Enabled,
			Example:     configuration.Example.Or(false),
			ContentType: cm.NewOptString(configuration.EntityId),
		})
	}

	return result
}

func validatePreviewEnvironmentData(data cm.PreviewEnvironmentData) *cm.ErrorStatusCode {
	seen := make(map[previewEnvironmentConfigurationIdentity]struct{}, len(data.Configurations))
	for _, configuration := range data.Configurations {
		if configuration.EntityType == "" || configuration.EntityId == "" {
			return NewContentfulManagementErrorStatusCode(
				http.StatusBadRequest,
				"ContentPreviewChangeInvalid",
				new("Configuration identity is required"),
				nil,
			)
		}

		configurationIdentity := previewEnvironmentConfigurationIdentity{
			EntityType: configuration.EntityType,
			EntityID:   configuration.EntityId,
		}
		if _, ok := seen[configurationIdentity]; ok {
			return NewContentfulManagementErrorStatusCode(
				http.StatusBadRequest,
				"ContentPreviewChangeInvalid",
				new("Duplicate configurations are not allowed"),
				nil,
			)
		}

		seen[configurationIdentity] = struct{}{}
	}

	return nil
}

type previewEnvironmentConfigurationIdentity struct {
	EntityType string
	EntityID   string
}
