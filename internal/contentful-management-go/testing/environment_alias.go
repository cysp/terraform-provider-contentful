package testing

import (
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewEnvironmentAliasFromEnvironmentAliasData(spaceID, environmentAliasID string, data cm.EnvironmentAliasData) cm.EnvironmentAlias {
	return cm.EnvironmentAlias{
		Sys:         cm.NewEnvironmentAliasSys(spaceID, environmentAliasID),
		Environment: data.Environment,
	}
}

func UpdateEnvironmentAliasFromEnvironmentAliasData(environmentAlias *cm.EnvironmentAlias, data cm.EnvironmentAliasData) {
	now := time.Now()

	environmentAlias.Sys.Version++
	environmentAlias.Sys.UpdatedAt = cm.NewOptDateTime(now)

	environmentAlias.Environment = data.Environment
}
