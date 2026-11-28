package provider_test

import (
	"archive/zip"
	"context"
	"io"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/require"
)

type afterPlan func()

func (f afterPlan) CheckPlan(context.Context, plancheck.CheckPlanRequest, *plancheck.CheckPlanResponse) {
	f()
}

func checkSavedPlanMembers(t *testing.T, filename string, check func(string)) {
	t.Helper()

	archive, err := zip.OpenReader(filename)
	require.NoError(t, err)

	defer func() { require.NoError(t, archive.Close()) }()

	for _, member := range archive.File {
		reader, err := member.Open()
		require.NoError(t, err)
		content, err := io.ReadAll(reader)
		require.NoError(t, err)
		require.NoError(t, reader.Close())
		check(string(content))
	}
}
