package testing

import (
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewEnvironment(spaceID, environmentID, name string) cm.Environment {
	now := time.Now()

	return cm.Environment{
		Sys: cm.EnvironmentSys{
			Type:      cm.EnvironmentSysTypeEnvironment,
			ID:        environmentID,
			Version:   1,
			CreatedAt: cm.NewOptDateTime(now),
			UpdatedAt: cm.NewOptDateTime(now),
			Space: cm.SpaceLink{
				Sys: cm.SpaceLinkSys{
					Type:     cm.SpaceLinkSysTypeLink,
					LinkType: cm.SpaceLinkSysLinkTypeSpace,
					ID:       spaceID,
				},
			},
		},
		Name: name,
	}
}

func NewEnvironmentFromRequest(spaceID, environmentID string, req cm.EnvironmentFields) cm.Environment {
	environment := NewEnvironment(spaceID, environmentID, req.Name)
	return environment
}

func UpdateEnvironmentFromRequest(environment *cm.Environment, req cm.EnvironmentFields) {
	now := time.Now()

	environment.Sys.Version++
	environment.Sys.UpdatedAt = cm.NewOptDateTime(now)

	environment.Name = req.Name
}
