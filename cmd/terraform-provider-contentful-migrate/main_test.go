package main_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	literal  = "resource \"contentful_entry\" \"example\" { fields = { title = jsonencode({ en = var.name }) } }\n"
	migrated = "resource \"contentful_entry\" \"example\" { fields = { title = { en = jsonencode(var.name) } } }\n"
)

func TestMigrationCommandWrites(t *testing.T) {
	t.Parallel()

	binary := filepath.Join(t.TempDir(), "migrate")
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binary, ".")
	output, err := build.CombinedOutput()
	require.NoError(t, err, "%s", output)

	t.Run("completed file before parse error", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		first, last := filepath.Join(dir, "a.tf"), filepath.Join(dir, "z.tf")
		require.NoError(t, os.WriteFile(first, []byte(literal), 0o600))
		require.NoError(t, os.WriteFile(last, []byte("resource {"), 0o600))

		stdout, stderr := runMigration(t, binary, 1, "-write", dir)
		assert.Equal(t, "updated: "+first+": localize fields on contentful_entry.example\n", stdout)
		assert.Contains(t, stderr, last)
		assert.Contains(t, stderr, "parse config")
		assertFileContents(t, first, migrated)
		assertFileContents(t, last, "resource {")
	})

	t.Run("symbolic link and repeated inputs", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		target, link := filepath.Join(dir, "z.tf"), filepath.Join(dir, "a.tf")
		require.NoError(t, os.WriteFile(target, []byte(literal), 0o600))

		const mode = os.FileMode(0o640)
		require.NoError(t, os.Chmod(target, mode))
		require.NoError(t, os.Symlink(filepath.Base(target), link))
		before, err := os.Stat(target)
		require.NoError(t, err)

		for _, step := range []struct {
			flags  []string
			exit   int
			action string
		}{
			{[]string{"-check"}, 2, "would update"},
			{[]string{"-write", "-check"}, 0, "updated"},
		} {
			stdout, stderr := runMigration(t, binary, step.exit, append(step.flags, dir, link, target)...)
			assert.Equal(t, step.action+": "+link+": localize fields on contentful_entry.example\n", stdout)
			assert.Empty(t, stderr)
		}

		name, err := os.Readlink(link)
		require.NoError(t, err)
		assert.Equal(t, filepath.Base(target), name)
		after, err := os.Stat(target)
		require.NoError(t, err)
		assert.Equal(t, mode, after.Mode().Perm())
		assert.False(t, os.SameFile(before, after))
		assertFileContents(t, target, migrated)
		assertFileContents(t, link, migrated)

		stdout, stderr := runMigration(t, binary, 0, "-write", "-check", dir)
		assert.Empty(t, stdout)
		assert.Empty(t, stderr)

		unchanged, err := os.Stat(target)
		require.NoError(t, err)
		assert.True(t, os.SameFile(after, unchanged))
	})

	t.Run("staging failure preserves file", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		target := filepath.Join(dir, "entry.tf")
		require.NoError(t, os.WriteFile(target, []byte(literal), 0o600))
		require.NoError(t, os.Chmod(dir, 0o500))
		t.Cleanup(func() { require.NoError(t, os.Chmod(dir, 0o700)) })

		probe, err := os.CreateTemp(dir, "write-access")
		if err == nil {
			require.NoError(t, probe.Close())
			t.Skip("current user can create files in the read-only directory")
		}

		require.ErrorIs(t, err, os.ErrPermission)

		stdout, stderr := runMigration(t, binary, 1, "-write", target)
		assert.Empty(t, stdout)
		assert.Contains(t, stderr, target)
		assertFileContents(t, target, literal)
	})

	t.Run("skipped and unwritable files", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		first, last := filepath.Join(dir, "a.tf"), filepath.Join(dir, "z.tf")

		const dynamic = "resource \"contentful_entry\" \"external\" { fields = var.fields }\n"
		require.NoError(t, os.WriteFile(first, []byte(dynamic), 0o600))
		require.NoError(t, os.WriteFile(last, []byte(literal), 0o400))

		writable, err := os.OpenFile(last, os.O_WRONLY, 0)
		if err == nil {
			require.NoError(t, writable.Close())
			t.Skip("current user can write the read-only fixture")
		}

		require.ErrorIs(t, err, os.ErrPermission)

		stdout, stderr := runMigration(t, binary, 1, "-write", "-check", dir)
		assert.Empty(t, stdout)
		assert.Contains(t, stderr, "skipped: "+first+": contentful_entry.external:")
		assert.Contains(t, stderr, last)
		assertFileContents(t, first, dynamic)
		assertFileContents(t, last, literal)
	})
}

func runMigration(t *testing.T, binary string, exit int, args ...string) (string, string) {
	t.Helper()

	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(t.Context(), binary, args...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	err := cmd.Run()
	if exit == 0 {
		require.NoError(t, err, "%s", &stderr)
	} else {
		var failure *exec.ExitError
		require.ErrorAs(t, err, &failure)
		require.Equal(t, exit, failure.ExitCode(), "%s", &stderr)
	}

	return stdout.String(), stderr.String()
}

func assertFileContents(t *testing.T, name, want string) {
	t.Helper()

	got, err := os.ReadFile(name)
	require.NoError(t, err)
	assert.Equal(t, want, string(got))
}
