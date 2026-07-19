package contentfulmanagement

func NewUserLink(id string) UserLink {
	return UserLink{Sys: UserLinkSys{
		Type:     UserLinkSysTypeLink,
		LinkType: UserLinkSysLinkTypeUser,
		ID:       id,
	}}
}
