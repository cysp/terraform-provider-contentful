package contentfulmanagement

type ErrorResponse interface {
	GetError() (v Error, ok bool)
}

type StatusCodeResponse interface {
	GetStatusCode() int
}

type ErrorStatusCodeResponse interface {
	GetStatusCode() int
	GetError() (v Error, ok bool)
}

func (r *ErrorStatusCode) GetError() (Error, bool) {
	return r.GetResponse().GetError()
}
