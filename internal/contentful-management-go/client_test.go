package client_test

func addressOf[T any](value T) *T {
	return &value
}
