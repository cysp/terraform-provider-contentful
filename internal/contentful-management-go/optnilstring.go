package client

func NewOptNilStringNull() OptNilString {
	return OptNilString{Set: true, Null: true}
}

func (v *OptNilString) ValueStringPointer() *string {
	if value, ok := v.Get(); ok {
		return &value
	}

	return nil
}
