package provider

import (
	"context"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *PreviewEnvironmentModel) ToPreviewEnvironmentData(_ context.Context, modelPath path.Path) (cm.PreviewEnvironmentData, diag.Diagnostics) {
	configurations, diagnostics := previewEnvironmentContentTypeConfigurationValues(model, modelPath)
	name, nameDiagnostics := requestRequiredString(model.Name, modelPath.AtName("name"))
	diagnostics.Append(nameDiagnostics...)

	description, descriptionDiagnostics := requestRequiredString(model.Description, modelPath.AtName("description"))
	diagnostics.Append(descriptionDiagnostics...)

	if diagnostics.HasError() {
		return cm.PreviewEnvironmentData{}, diagnostics
	}

	keys := sortedPreviewEnvironmentConfigurationKeys(configurations)

	requestConfigurations := make([]cm.PreviewEnvironmentConfigurationData, 0, len(keys))
	for _, contentTypeID := range keys {
		requestConfigurations = append(requestConfigurations, newPreviewEnvironmentConfigurationData(
			contentTypeID,
			configurations[contentTypeID].URL.ValueString(),
			true,
		))
	}

	return newPreviewEnvironmentData(name, description, requestConfigurations), diagnostics
}

func ToPreviewEnvironmentUpdateData(
	_ context.Context,
	modelPath path.Path,
	state *PreviewEnvironmentModel,
	plan *PreviewEnvironmentModel,
) (cm.PreviewEnvironmentData, diag.Diagnostics) {
	stateConfigurations, diagnostics := previewEnvironmentContentTypeConfigurationValues(state, modelPath)

	planConfigurations, planDiagnostics := previewEnvironmentContentTypeConfigurationValues(plan, modelPath)
	diagnostics.Append(planDiagnostics...)

	name, nameDiagnostics := requestRequiredString(plan.Name, modelPath.AtName("name"))
	diagnostics.Append(nameDiagnostics...)

	description, descriptionDiagnostics := requestRequiredString(plan.Description, modelPath.AtName("description"))
	diagnostics.Append(descriptionDiagnostics...)

	if diagnostics.HasError() {
		return cm.PreviewEnvironmentData{}, diagnostics
	}

	requestConfigurations := make([]cm.PreviewEnvironmentConfigurationData, 0)

	for _, contentTypeID := range sortedPreviewEnvironmentConfigurationKeys(planConfigurations) {
		planConfiguration := planConfigurations[contentTypeID]

		stateConfiguration, existed := stateConfigurations[contentTypeID]
		if existed && stateConfiguration.URL.Equal(planConfiguration.URL) {
			continue
		}

		requestConfigurations = append(requestConfigurations, newPreviewEnvironmentConfigurationData(
			contentTypeID,
			planConfiguration.URL.ValueString(),
			true,
		))
	}

	for _, contentTypeID := range sortedPreviewEnvironmentConfigurationKeys(stateConfigurations) {
		if _, remainsEnabled := planConfigurations[contentTypeID]; remainsEnabled {
			continue
		}

		requestConfigurations = append(requestConfigurations, newPreviewEnvironmentConfigurationData(
			contentTypeID,
			stateConfigurations[contentTypeID].URL.ValueString(),
			false,
		))
	}

	return newPreviewEnvironmentData(name, description, requestConfigurations), diagnostics
}

func previewEnvironmentContentTypeConfigurationValues(
	model *PreviewEnvironmentModel,
	modelPath path.Path,
) (map[string]PreviewEnvironmentContentTypeConfigurationValue, diag.Diagnostics) {
	diagnostics := diag.Diagnostics{}

	configurationsPath := modelPath.AtName("content_type_configurations")
	if model.ContentTypeConfigurations.IsNull() || model.ContentTypeConfigurations.IsUnknown() {
		diagnostics.AddAttributeError(
			configurationsPath,
			"Invalid content type preview configurations",
			"Content type configurations must be known and non-null.",
		)

		return nil, diagnostics
	}

	configurations := make(map[string]PreviewEnvironmentContentTypeConfigurationValue, len(model.ContentTypeConfigurations.Elements()))
	for contentTypeID, configurationObject := range model.ContentTypeConfigurations.Elements() {
		configurationPath := configurationsPath.AtMapKey(contentTypeID)

		configuration, ok := configurationObject.GetValue()
		if !ok {
			diagnostics.AddAttributeError(
				configurationPath,
				"Invalid content type preview configuration",
				"Content type configuration must be known and non-null.",
			)

			continue
		}

		if configuration.URL.IsNull() || configuration.URL.IsUnknown() {
			diagnostics.AddAttributeError(
				configurationPath.AtName("url"),
				"Invalid preview URL",
				"Preview URL must be known and non-null.",
			)

			continue
		}

		configurations[contentTypeID] = configuration
	}

	return configurations, diagnostics
}

func sortedPreviewEnvironmentConfigurationKeys(
	configurations map[string]PreviewEnvironmentContentTypeConfigurationValue,
) []string {
	keys := make([]string, 0, len(configurations))
	for key := range configurations {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	return keys
}

func newPreviewEnvironmentConfigurationData(contentTypeID, url string, enabled bool) cm.PreviewEnvironmentConfigurationData {
	return cm.PreviewEnvironmentConfigurationData{
		URL:        url,
		EntityType: "ContentType",
		EntityId:   contentTypeID,
		Enabled:    enabled,
	}
}

func newPreviewEnvironmentData(
	name string,
	description string,
	configurations []cm.PreviewEnvironmentConfigurationData,
) cm.PreviewEnvironmentData {
	return cm.PreviewEnvironmentData{
		Name:           name,
		Description:    description,
		Configurations: configurations,
	}
}
