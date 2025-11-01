package testing

import (
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewEnvironmentAliasSys(spaceID, environmentAliasID string) cm.EnvironmentAliasSys {
	now := time.Now()

	return cm.EnvironmentAliasSys{
		Type:      cm.EnvironmentAliasSysTypeEnvironmentAlias,
		ID:        environmentAliasID,
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
	}
}

func NewEnvironmentAliasFromEnvironmentAliasData(spaceID, environmentAliasID string, data cm.EnvironmentAliasData) cm.EnvironmentAlias {
	return cm.EnvironmentAlias{
		Sys:         NewEnvironmentAliasSys(spaceID, environmentAliasID),
		Environment: data.Environment,
	}
}

func UpdateEnvironmentAliasFromEnvironmentAliasData(environmentAlias *cm.EnvironmentAlias, data cm.EnvironmentAliasData) {
	now := time.Now()

	environmentAlias.Sys.Version++
	environmentAlias.Sys.UpdatedAt = cm.NewOptDateTime(now)

	environmentAlias.Environment = data.Environment
}
