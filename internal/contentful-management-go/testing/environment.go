package testing

import (
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewEnvironmentFromEnvironmentData(spaceID, environmentID string, data cm.EnvironmentData) cm.Environment {
	now := time.Now()

	return cm.Environment{
		Sys:  cm.NewEnvironmentSys(spaceID, environmentID, now),
		Name: data.Name,
	}
}

func UpdateEnvironmentFromEnvironmentData(environment *cm.Environment, data cm.EnvironmentData) {
	now := time.Now()

	environment.Sys.Version++
	environment.Sys.UpdatedAt = cm.NewOptDateTime(now)

	environment.Name = data.Name
}
