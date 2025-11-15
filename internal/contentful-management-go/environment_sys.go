package contentfulmanagement

import "time"

func NewEnvironmentSys(spaceID, environmentID string, createdAt time.Time) EnvironmentSys {
	return EnvironmentSys{
		Type:      EnvironmentSysTypeEnvironment,
		ID:        environmentID,
		Version:   1,
		CreatedAt: NewOptDateTime(createdAt),
		UpdatedAt: NewOptDateTime(createdAt),
		Space:     NewSpaceLink(spaceID),
	}
}
