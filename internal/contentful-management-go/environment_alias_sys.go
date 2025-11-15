package contentfulmanagement

import (
	"time"
)

func NewEnvironmentAliasSys(spaceID, environmentAliasID string) EnvironmentAliasSys {
	now := time.Now()

	return EnvironmentAliasSys{
		Type:      EnvironmentAliasSysTypeEnvironmentAlias,
		ID:        environmentAliasID,
		Version:   1,
		CreatedAt: NewOptDateTime(now),
		UpdatedAt: NewOptDateTime(now),
		Space:     NewSpaceLink(spaceID),
	}
}
