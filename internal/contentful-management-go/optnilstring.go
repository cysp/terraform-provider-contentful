package contentfulmanagement

func NewOptNilStringNull() OptNilString {
	return OptNilString{Set: true, Null: true}
}

func NewOptNilPointerString(v *string) OptNilString {
	if v == nil {
		return NewOptNilStringNull()
	}

	return NewOptNilString(*v)
}

func (v *OptNilString) ValueStringPointer() *string {
	if value, ok := v.Get(); ok {
		return &value
	}

	return nil
}
