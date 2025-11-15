package contentfulmanagement

func NewOrganizationLink(id string) OrganizationLink {
	return OrganizationLink{
		Sys: NewOrganizationLinkSys(id),
	}
}

func NewOrganizationLinkSys(id string) OrganizationLinkSys {
	return OrganizationLinkSys{
		Type:     OrganizationLinkSysTypeLink,
		LinkType: OrganizationLinkSysLinkTypeOrganization,
		ID:       id,
	}
}
