package testing

import (
	"math/rand"
)

const (
	NonexistentID = "nonexistent"
)

func generateResourceID() string {
	return RandStringBytes(8) //nolint:mnd
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	bytes := make([]byte, n)

	for i := range bytes {
		//nolint:gosec
		bytes[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(bytes)
}
