package contentfulmanagement

func NewRoleSys(spaceID, roleID string) RoleSys {
	return RoleSys{
		Type:  RoleSysTypeRole,
		Space: NewSpaceLink(spaceID),
		ID:    roleID,
	}
}
