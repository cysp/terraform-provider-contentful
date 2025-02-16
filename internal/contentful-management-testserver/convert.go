package contentfulmanagementtestserver

type ConstraintOptNil[T any] interface {
	Get() (T, bool)
	SetTo(value T)
	Reset()
}

func convertOptNil[I any, O any](o ConstraintOptNil[O], i ConstraintOptNil[I], f func(I) O) {
	if value, ok := i.Get(); ok {
		o.SetTo(f(value))
	} else {
		o.Reset()
	}
}

func convertSlice[I any, O any](i []I, f func(I) O) []O {
	out := make([]O, len(i))

	for index, item := range i {
		out[index] = f(item)
	}

	return out
}

func convertMap[I any, O any](i map[string]I, f func(I) O) map[string]O {
	out := make(map[string]O, len(i))

	for key, item := range i {
		out[key] = f(item)
	}

	return out
}
