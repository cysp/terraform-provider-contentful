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

func (r *ApplicationJSONErrorStatusCode) GetError() (Error, bool) {
	return r.GetResponse().GetError()
}

func (r *ApplicationVndContentfulManagementV1JSONErrorStatusCode) GetError() (Error, bool) {
	return r.GetResponse().GetError()
}
