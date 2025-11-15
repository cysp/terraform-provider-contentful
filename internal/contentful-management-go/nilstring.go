package contentfulmanagement

func NewNilStringNull() NilString {
	return NilString{Null: true}
}

func NewNilPointerString(v *string) NilString {
	if v == nil {
		return NewNilStringNull()
	}

	return NewNilString(*v)
}

func (v *NilString) ValueStringPointer() *string {
	if value, ok := v.Get(); ok {
		return &value
	}

	return nil
}
