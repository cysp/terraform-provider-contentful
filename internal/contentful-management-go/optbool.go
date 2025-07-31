package contentfulmanagement

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
