package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatVersionReleaseBuild(t *testing.T) {
	t.Parallel()

	expected := "terraform-provider-contentful 0.0.63 (commit 0123456789abcdef0123456789abcdef01234567)\n"
	actual := formatVersion("0.0.63", "0123456789abcdef0123456789abcdef01234567")

	assert.Equal(t, expected, actual)
}

func TestFormatVersionDevelopmentBuild(t *testing.T) {
	t.Parallel()

	expected := "terraform-provider-contentful dev (commit unknown)\n"
	actual := formatVersion("dev", "unknown")

	assert.Equal(t, expected, actual)
}
