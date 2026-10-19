// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package files

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
)

func WithSuffix(suffix string, paths ...string) ([]string, error) {
	var found []string
	errs := make([]error, len(paths))
	for index, path := range paths {
		var matches []string
		matches, errs[index] = filesUnder(path, suffix)
		found = append(found, matches...)
	}
	return found, errors.Join(errs...)
}

func filesUnder(root, suffix string) ([]string, error) {
	resolved, err := filepath.EvalSymlinks(root)
	if err == nil {
		return matchesIn(root, resolved, suffix)
	} else {
		return nil, err
	}
}

func matchesIn(root, resolved, suffix string) ([]string, error) {
	var matches []string
	err := filepath.WalkDir(resolved, func(path string, entry fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case path != resolved && entry.IsDir() && strings.HasPrefix(entry.Name(), "."):
			return filepath.SkipDir
		case !entry.IsDir() && strings.HasSuffix(path, suffix):
			matches = append(matches, filepath.Join(root, strings.TrimPrefix(path, resolved)))
			return nil
		default:
			return nil
		}
	})
	return matches, err
}
