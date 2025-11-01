package testing

import (
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewEnvironmentAlias(spaceID, environmentAliasID, targetEnvironmentID string) cm.EnvironmentAlias {
	return cm.EnvironmentAlias{
		Sys: NewEnvironmentAliasSys(spaceID, environmentAliasID),
		Environment: cm.EnvironmentLink{
			Sys: cm.EnvironmentLinkSys{
				Type:     cm.EnvironmentLinkSysTypeLink,
				LinkType: cm.EnvironmentLinkSysLinkTypeEnvironment,
				ID:       targetEnvironmentID,
			},
		},
	}
}

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

func NewEnvironmentAliasFromRequest(spaceID, environmentAliasID string, req cm.EnvironmentAliasRequest) cm.EnvironmentAlias {
	return cm.EnvironmentAlias{
		Sys:         NewEnvironmentAliasSys(spaceID, environmentAliasID),
		Environment: req.Environment,
	}
}

func UpdateEnvironmentAliasFromRequest(environmentAlias *cm.EnvironmentAlias, req cm.EnvironmentAliasRequest) {
	now := time.Now()

	environmentAlias.Sys.Version++
	environmentAlias.Sys.UpdatedAt = cm.NewOptDateTime(now)

	environmentAlias.Environment = req.Environment
}
