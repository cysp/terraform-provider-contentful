package testing

import (
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewEnvironmentSys(spaceID, environmentID string, createdAt time.Time) cm.EnvironmentSys {
	return cm.EnvironmentSys{
		Type:      cm.EnvironmentSysTypeEnvironment,
		ID:        environmentID,
		Version:   1,
		CreatedAt: cm.NewOptDateTime(createdAt),
		UpdatedAt: cm.NewOptDateTime(createdAt),
		Space: cm.SpaceLink{
			Sys: cm.SpaceLinkSys{
				Type:     cm.SpaceLinkSysTypeLink,
				LinkType: cm.SpaceLinkSysLinkTypeSpace,
				ID:       spaceID,
			},
		},
	}
}

func NewEnvironmentFromEnvironmentData(spaceID, environmentID string, data cm.EnvironmentData) cm.Environment {
	now := time.Now()

	return cm.Environment{
		Sys:  NewEnvironmentSys(spaceID, environmentID, now),
		Name: data.Name,
	}
}

func UpdateEnvironmentFromEnvironmentData(environment *cm.Environment, data cm.EnvironmentData) {
	now := time.Now()

	environment.Sys.Version++
	environment.Sys.UpdatedAt = cm.NewOptDateTime(now)

	environment.Name = data.Name
}
