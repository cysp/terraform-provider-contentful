package migrate

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Options struct {
	Paths []string
	Write bool
}

type Report struct {
	Changes []Change
	Skips   []Skip
}

type Change struct {
	Path        string
	Description string
}

type Skip struct {
	Path    string
	Address string
	Reason  string
}

type fileReport struct {
	Changes []Change
	Skips   []Skip
}

type migration interface {
	Apply(path string, src []byte) ([]byte, fileReport, error)
}

func Run(ctx context.Context, options Options) (Report, error) {
	if len(options.Paths) == 0 {
		options.Paths = []string{"."}
	}

	migrations := []migration{
		entryFieldsMigration{},
	}

	files, err := terraformFiles(options.Paths)
	if err != nil {
		return Report{}, err
	}

	report := Report{}

	for _, path := range files {
		err := ctx.Err()
		if err != nil {
			return report, fmt.Errorf("migration cancelled: %w", err)
		}

		original, err := os.ReadFile(path)
		if err != nil {
			return report, fmt.Errorf("read %s: %w", path, err)
		}

		next := slices.Clone(original)
		fileReport := fileReport{}

		for _, migration := range migrations {
			migrated, migrationReport, err := migration.Apply(path, next)
			if err != nil {
				return report, fmt.Errorf("apply migration to %s: %w", path, err)
			}

			next = migrated

			fileReport.Changes = append(fileReport.Changes, migrationReport.Changes...)
			fileReport.Skips = append(fileReport.Skips, migrationReport.Skips...)
		}

		report.Changes = append(report.Changes, fileReport.Changes...)
		report.Skips = append(report.Skips, fileReport.Skips...)

		if options.Write && !bytes.Equal(original, next) {
			info, err := os.Stat(path)
			if err != nil {
				return report, fmt.Errorf("stat %s: %w", path, err)
			}

			//nolint:gosec // This CLI intentionally writes user-selected Terraform files.
			err = os.WriteFile(path, next, info.Mode())
			if err != nil {
				return report, fmt.Errorf("write %s: %w", path, err)
			}
		}
	}

	return report, nil
}

func terraformFiles(paths []string) ([]string, error) {
	var files []string

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("stat %s: %w", path, err)
		}

		if !info.IsDir() {
			if strings.HasSuffix(path, ".tf") {
				files = append(files, path)
			}

			continue
		}

		err = filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if entry.IsDir() {
				switch entry.Name() {
				case ".git", ".terraform":
					return filepath.SkipDir
				}

				return nil
			}

			if strings.HasSuffix(path, ".tf") {
				files = append(files, path)
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk %s: %w", path, err)
		}
	}

	slices.Sort(files)
	files = slices.Compact(files)

	return files, nil
}
