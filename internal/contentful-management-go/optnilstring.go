package client

func NewOptNilPointerString(value *string) OptNilString {
	if value == nil {
		return OptNilString{}
	}

	return NewOptNilString(*value)
}

func (v *OptNilString) ValueStringPointer() *string {
	if value, ok := v.Get(); ok {
		return &value
	}

	return nil
}
