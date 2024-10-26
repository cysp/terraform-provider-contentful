package client

func NewOptNilIntNull() OptNilInt {
	return OptNilInt{Set: true, Null: true}
}

func NewOptNilPointerInt(value *int) OptNilInt {
	if value == nil {
		return NewOptNilIntNull()
	}

	return NewOptNilInt(*value)
}

func NewOptNilPointerInt64(value *int64) OptNilInt {
	if value == nil {
		return NewOptNilIntNull()
	}

	return NewOptNilInt(int(*value))
}

func (v *OptNilInt) ValueIntPointer() *int {
	if value, ok := v.Get(); ok {
		return &value
	}

	return nil
}

func (v *OptNilInt) ValueInt64Pointer() *int64 {
	if value, ok := v.Get(); ok {
		value := int64(value)

		return &value
	}

	return nil
}
