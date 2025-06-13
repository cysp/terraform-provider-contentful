package contentfulmanagementtestserver

type ConstraintOptNil[T any] interface {
	Get() (T, bool)
	SetTo(value T)
	Reset()
}

func convertOptNil[I any, O any](o ConstraintOptNil[O], i ConstraintOptNil[I], convert func(I) O) {
	if value, ok := i.Get(); ok {
		o.SetTo(convert(value))
	} else {
		o.Reset()
	}
}

func convertSlice[I any, O any](i []I, convert func(I) O) []O {
	if i == nil {
		return nil
	}

	out := make([]O, len(i))

	for index, item := range i {
		out[index] = convert(item)
	}

	return out
}

func convertMap[I any, O any](i map[string]I, convert func(I) O) map[string]O {
	if i == nil {
		return nil
	}

	out := make(map[string]O, len(i))

	for key, item := range i {
		out[key] = convert(item)
	}

	return out
}
