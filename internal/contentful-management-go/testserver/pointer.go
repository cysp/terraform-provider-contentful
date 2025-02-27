package testserver

func pointerTo[T any](value T) *T {
	return &value
}
