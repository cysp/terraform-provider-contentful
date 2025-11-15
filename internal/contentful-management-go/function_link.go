package contentfulmanagement

func NewFunctionLink(id string) FunctionLink {
	return FunctionLink{
		Sys: NewFunctionLinkSys(id),
	}
}

func NewFunctionLinkSys(id string) FunctionLinkSys {
	return FunctionLinkSys{
		Type:     FunctionLinkSysTypeLink,
		LinkType: FunctionLinkSysLinkTypeFunction,
		ID:       id,
	}
}
