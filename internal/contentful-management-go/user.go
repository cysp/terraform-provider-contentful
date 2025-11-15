package contentfulmanagement

func NewUser(id string) *User {
	return &User{
		Sys: NewUserSys(id),
	}
}
