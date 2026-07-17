package provider

import (
	"context"
	"fmt"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *PreviewEnvironmentModel) ToPreviewEnvironmentData(_ context.Context, modelPath path.Path) (cm.PreviewEnvironmentData, diag.Diagnostics) {
	configurations, diagnostics := previewEnvironmentContentTypeConfigurationValues(model, modelPath)
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

	return model.newPreviewEnvironmentData(requestConfigurations), diagnostics
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

	return plan.newPreviewEnvironmentData(requestConfigurations), diagnostics
}

func ValidatePreviewEnvironmentUpdateResponse(
	_ context.Context,
	modelPath path.Path,
	state *PreviewEnvironmentModel,
	plan *PreviewEnvironmentModel,
	response *PreviewEnvironmentModel,
) diag.Diagnostics {
	stateConfigurations, diagnostics := previewEnvironmentContentTypeConfigurationValues(state, modelPath)

	planConfigurations, planDiagnostics := previewEnvironmentContentTypeConfigurationValues(plan, modelPath)
	diagnostics.Append(planDiagnostics...)

	responseConfigurations, responseDiagnostics := previewEnvironmentContentTypeConfigurationValues(response, modelPath)
	diagnostics.Append(responseDiagnostics...)

	if diagnostics.HasError() {
		return diagnostics
	}

	configurationsPath := modelPath.AtName("content_type_configurations")

	for contentTypeID, planConfiguration := range planConfigurations {
		stateConfiguration, existed := stateConfigurations[contentTypeID]
		if existed && stateConfiguration.URL.Equal(planConfiguration.URL) {
			continue
		}

		responseConfiguration, enabled := responseConfigurations[contentTypeID]
		if !enabled || !responseConfiguration.URL.Equal(planConfiguration.URL) {
			diagnostics.AddAttributeError(
				configurationsPath.AtMapKey(contentTypeID),
				"Content preview configuration update was not applied",
				fmt.Sprintf("Contentful did not return content type %q enabled with the planned URL.", contentTypeID),
			)
		}
	}

	for contentTypeID := range stateConfigurations {
		if _, remainsEnabled := planConfigurations[contentTypeID]; remainsEnabled {
			continue
		}

		if _, remainsEnabled := responseConfigurations[contentTypeID]; remainsEnabled {
			diagnostics.AddAttributeError(
				configurationsPath.AtMapKey(contentTypeID),
				"Content preview configuration removal was not applied",
				fmt.Sprintf("Contentful returned content type %q still enabled after it was disabled.", contentTypeID),
			)
		}
	}

	return diagnostics
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

func (model *PreviewEnvironmentModel) newPreviewEnvironmentData(
	configurations []cm.PreviewEnvironmentConfigurationData,
) cm.PreviewEnvironmentData {
	return cm.PreviewEnvironmentData{
		Name:           model.Name.ValueString(),
		Description:    model.Description.ValueString(),
		Configurations: configurations,
	}
}
