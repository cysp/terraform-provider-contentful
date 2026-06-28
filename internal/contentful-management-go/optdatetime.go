package contentfulmanagement

import "time"

func (v *OptDateTime) ValueTimePointer() *time.Time {
	if value, ok := v.Get(); ok {
		return &value
	}

	return nil
}
