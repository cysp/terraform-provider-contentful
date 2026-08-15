package migrate_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/migrate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunMigratesContentfulEntryFields(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "entry.tf")
	require.NoError(t, os.WriteFile(path, []byte(`
resource "contentful_entry" "example" {
  space_id        = var.space_id
  environment_id  = var.environment_id
  content_type_id = "blogPost"

  fields = {
    title = jsonencode({
      "en-AU" = "My First Blog Post"
      "en-US" = upper(var.title)
    })
    body = jsonencode({
      "en-AU" = {
        nodeType = "document"
        data     = {}
        content  = []
      }
    })
  }
}
`), 0o600))

	report, err := migrate.Run(context.Background(), migrate.Options{
		Paths: []string{path},
		Write: true,
	})
	require.NoError(t, err)
	require.Len(t, report.Changes, 1)
	assert.Empty(t, report.Skips)

	gotBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	got := string(gotBytes)

	assert.Contains(t, got, `"en-AU" = jsonencode("My First Blog Post")`)
	assert.Contains(t, got, `"en-US" = jsonencode(upper(var.title))`)
	assert.Contains(t, got, `body = {`)
	assert.NotContains(t, got, `title = jsonencode({`)
}

func TestRunDryRunDoesNotWrite(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "entry.tf")
	original := []byte(`
resource "contentful_entry" "example" {
  fields = {
    title = jsonencode({
      "en-AU" = "Title"
    })
  }
}
`)
	require.NoError(t, os.WriteFile(path, original, 0o600))

	report, err := migrate.Run(context.Background(), migrate.Options{Paths: []string{path}})
	require.NoError(t, err)
	require.Len(t, report.Changes, 1)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, string(original), string(got))
}

func TestRunSkipsDynamicFields(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "entry.tf")
	require.NoError(t, os.WriteFile(path, []byte(`
resource "contentful_entry" "example" {
  fields = var.fields
}
`), 0o600))

	report, err := migrate.Run(context.Background(), migrate.Options{
		Paths: []string{path},
		Write: true,
	})
	require.NoError(t, err)
	assert.Empty(t, report.Changes)
	require.Len(t, report.Skips, 1)
	assert.Equal(t, "contentful_entry.example", report.Skips[0].Address)
	assert.Contains(t, report.Skips[0].Reason, "literal object")

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(got), "fields = var.fields")
}

func TestRunSkipsPartiallyMigratableResource(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "entry.tf")
	original := `
resource "contentful_entry" "example" {
  fields = {
    title = jsonencode({
      "en-AU" = "Title"
    })
    body = var.body
  }
}
`
	require.NoError(t, os.WriteFile(path, []byte(original), 0o600))

	report, err := migrate.Run(context.Background(), migrate.Options{
		Paths: []string{path},
		Write: true,
	})
	require.NoError(t, err)
	assert.Empty(t, report.Changes)
	require.Len(t, report.Skips, 1)
	assert.Contains(t, report.Skips[0].Reason, "fields.body")

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(original), strings.TrimSpace(string(got)))
}

func TestRunDoesNotReportAlreadyMigratedFields(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "entry.tf")
	original := `
resource "contentful_entry" "example" {
  fields = {
    title = {
      "en-AU" = jsonencode("Title")
    }
  }
}
`
	require.NoError(t, os.WriteFile(path, []byte(original), 0o600))

	report, err := migrate.Run(context.Background(), migrate.Options{
		Paths: []string{path},
		Write: true,
	})
	require.NoError(t, err)
	assert.Empty(t, report.Changes)
	assert.Empty(t, report.Skips)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(original), strings.TrimSpace(string(got)))
}

func TestRunMigratesMixedOldAndNewLiteralFields(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "entry.tf")
	require.NoError(t, os.WriteFile(path, []byte(`
resource "contentful_entry" "example" {
  fields = {
    title = jsonencode({
      "en-AU" = "Title"
    })
    slug = {
      "en-AU" = jsonencode("title")
    }
  }
}
`), 0o600))

	report, err := migrate.Run(context.Background(), migrate.Options{
		Paths: []string{path},
		Write: true,
	})
	require.NoError(t, err)
	require.Len(t, report.Changes, 1)
	assert.Empty(t, report.Skips)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(got), `"en-AU" = jsonencode("Title")`)
	assert.Contains(t, string(got), `"en-AU" = jsonencode("title")`)
	assert.NotContains(t, string(got), `title = jsonencode({`)
}
