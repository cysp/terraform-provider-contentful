package contentfulmanagement

func NewEditorInterfaceSys(spaceID, environmentID, contentTypeID, id string) EditorInterfaceSys {
	return EditorInterfaceSys{
		Type:        EditorInterfaceSysTypeEditorInterface,
		ID:          id,
		Space:       NewSpaceLink(spaceID),
		Environment: NewEnvironmentLink(environmentID),
		ContentType: NewContentTypeLink(contentTypeID),
	}
}
