package contentfulmanagement

func NewPreviewEnvironmentCreateData(data PreviewEnvironmentData) PreviewEnvironmentCreateData {
	configurations := make([]PreviewEnvironmentCreateConfigurationData, 0, len(data.Configurations))
	for _, configuration := range data.Configurations {
		configurations = append(configurations, PreviewEnvironmentCreateConfigurationData{
			URL:        configuration.URL,
			EntityType: NewOptString(configuration.EntityType),
			EntityId:   NewOptString(configuration.EntityId),
			Enabled:    configuration.Enabled,
			Example:    configuration.Example,
		})
	}

	return PreviewEnvironmentCreateData{
		Name:           data.Name,
		Description:    data.Description,
		Configurations: configurations,
	}
}

func NewPreviewEnvironmentDataFromCreate(data PreviewEnvironmentCreateData) PreviewEnvironmentData {
	configurations := make([]PreviewEnvironmentConfigurationData, 0, len(data.Configurations))
	for _, configuration := range data.Configurations {
		entityID := configuration.EntityId.Or(configuration.ContentType.Or(""))

		configurations = append(configurations, PreviewEnvironmentConfigurationData{
			URL:        configuration.URL,
			EntityType: configuration.EntityType.Or("ContentType"),
			EntityId:   entityID,
			Enabled:    configuration.Enabled,
			Example:    configuration.Example,
		})
	}

	return PreviewEnvironmentData{
		Name:           data.Name,
		Description:    data.Description,
		Configurations: configurations,
	}
}
