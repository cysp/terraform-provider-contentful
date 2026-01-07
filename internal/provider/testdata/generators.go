package testdata

import "pgregory.net/rapid"

var alphanumericRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func AlphanumericRune() *rapid.Generator[rune] {
	return rapid.RuneFrom(alphanumericRunes)
}

func AlphanumericStringOfN(minLen, maxLen int) *rapid.Generator[string] {
	return rapid.StringOfN(AlphanumericRune(), minLen, maxLen, maxLen)
}
