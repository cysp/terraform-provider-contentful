package client

func NewOptPointerBool(value *bool) OptBool {
	if value == nil {
		return OptBool{}
	}

	return NewOptBool(*value)
}

func (v *OptBool) ValueBoolPointer() *bool {
	if value, ok := v.Get(); ok {
		return &value
	}

	return nil
}

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
