package client

import "time"

func NewOptNilDateTimeNull() OptNilDateTime {
	return OptNilDateTime{Set: true, Null: true}
}

func (v *OptNilDateTime) ValueTimePointer() *time.Time {
	if value, ok := v.Get(); ok {
		return &value
	}

	return nil
}
