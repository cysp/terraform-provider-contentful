package contentfulmanagement

func NewUserSys(id string) UserSys {
	return UserSys{
		Type: UserSysTypeUser,
		ID:   id,
	}
}
