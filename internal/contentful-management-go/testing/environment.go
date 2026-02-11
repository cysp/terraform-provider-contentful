package testing

import cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"

func NewEnvironmentFromEnvironmentData(spaceID, environmentID, status string, data cm.EnvironmentData) cm.Environment {
	return cm.Environment{
		Sys:  cm.NewEnvironmentSys(spaceID, environmentID, status),
		Name: data.Name,
	}
}

func UpdateEnvironmentFromEnvironmentData(environment *cm.Environment, data cm.EnvironmentData) {
	environment.Sys.Version++

	environment.Name = data.Name
}
