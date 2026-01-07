package testdata

import (
	"pgregory.net/rapid"
)

func RandomZeroable[T any](elem *rapid.Generator[T]) *rapid.Generator[T] {
	return zeroable(
		rapid.Bool(),
		elem,
	)
}

func zeroable[T any](present *rapid.Generator[bool], elem *rapid.Generator[T]) *rapid.Generator[T] {
	return rapid.Custom(func(t *rapid.T) T {
		if !present.Draw(t, "present") {
			var zero T
			return zero
		}

		return elem.Draw(t, "value")
	})
}
