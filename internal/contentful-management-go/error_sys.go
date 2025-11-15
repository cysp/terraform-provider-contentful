package contentfulmanagement

func NewErrorSys(id string) ErrorSys {
	return ErrorSys{
		Type: ErrorSysTypeError,
		ID:   id,
	}
}
