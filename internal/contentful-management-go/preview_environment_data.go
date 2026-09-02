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
