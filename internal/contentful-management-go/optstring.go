package client

func NewOptPointerString(value *string) OptString {
	if value == nil {
		return OptString{}
	}

	return NewOptString(*value)
}

func (v *OptString) ValueStringPointer() *string {
	if value, ok := v.Get(); ok {
		return &value
	}

	return nil
}
