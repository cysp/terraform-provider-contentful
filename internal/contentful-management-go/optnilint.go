package contentfulmanagement

func NewOptNilIntNull() OptNilInt {
	return OptNilInt{Set: true, Null: true}
}

func NewOptNilPointerInt64(value *int64) OptNilInt {
	if value == nil {
		return NewOptNilIntNull()
	}

	return NewOptNilInt(int(*value))
}
