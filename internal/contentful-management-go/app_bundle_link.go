package contentfulmanagement

func NewAppBundleLink(id string) AppBundleLink {
	return AppBundleLink{
		Sys: NewAppBundleLinkSys(id),
	}
}

func NewAppBundleLinkSys(id string) AppBundleLinkSys {
	return AppBundleLinkSys{
		Type:     AppBundleLinkSysTypeLink,
		LinkType: AppBundleLinkSysLinkTypeAppBundle,
		ID:       id,
	}
}
