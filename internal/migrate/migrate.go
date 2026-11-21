package migrate

import (
	"bytes"
	"context"
	"errors"
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

func Run(ctx context.Context, options Options) (Report, error) {
	if len(options.Paths) == 0 {
		options.Paths = []string{"."}
	}

	files, err := terraformFiles(options.Paths)
	if err != nil {
		return Report{}, err
	}

	report := Report{}
	visited := make(map[string]struct{})

	for _, path := range files {
		err := ctx.Err()
		if err != nil {
			return report, fmt.Errorf("migration cancelled: %w", err)
		}

		absolute, err := filepath.Abs(path)
		if err != nil {
			return report, fmt.Errorf("resolve %s: %w", path, err)
		}

		target, err := filepath.EvalSymlinks(absolute)
		if err != nil {
			return report, fmt.Errorf("resolve %s: %w", path, err)
		}

		if _, ok := visited[target]; ok {
			continue
		}

		visited[target] = struct{}{}

		original, err := os.ReadFile(target)
		if err != nil {
			return report, fmt.Errorf("read %s: %w", path, err)
		}

		next, fileReport, err := migrateEntryFields(path, original)
		if err != nil {
			return report, fmt.Errorf("migrate %s: %w", path, err)
		}

		report.Skips = append(report.Skips, fileReport.Skips...)

		if options.Write && !bytes.Equal(original, next) {
			err = replaceFile(target, next)
			if err != nil {
				return report, fmt.Errorf("write %s: %w", path, err)
			}
		}

		report.Changes = append(report.Changes, fileReport.Changes...)
	}

	return report, nil
}

// Stage the complete replacement beside the resolved target. Opening the target
// for writing without truncation keeps its write permissions authoritative.
func replaceFile(target string, content []byte) error {
	original, err := os.OpenFile(target, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("open target: %w", err)
	}

	info, statErr := original.Stat()

	err = errors.Join(statErr, original.Close())
	if err != nil {
		return fmt.Errorf("inspect target: %w", err)
	}

	staged, err := os.CreateTemp(filepath.Dir(target), ".contentful-migrate-*")
	if err != nil {
		return fmt.Errorf("create temporary file: %w", err)
	}

	_, err = staged.Write(content)
	if err == nil {
		err = staged.Chmod(info.Mode().Perm())
	}

	err = errors.Join(err, staged.Close())
	if err == nil {
		err = os.Rename(staged.Name(), target)
	}

	if err != nil {
		return fmt.Errorf("replace target: %w", errors.Join(err, os.Remove(staged.Name())))
	}

	return nil
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
