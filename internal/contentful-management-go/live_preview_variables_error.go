package contentfulmanagement

func (r *LivePreviewVariablesErrorStatusCode) GetError() (Error, bool) {
	return r.Response.GetError()
}
