// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package directories

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/thehale/tooluse-screener/internal/lists"
)

func AllUnder(paths, roots []string) bool {
	return lists.Every(paths, func(path string) bool {
		return lists.Some(roots, func(root string) bool { return contains(root, path) })
	})
}

func contains(root, path string) bool {
	outer, outerKnown := absolute(root)
	inner, innerKnown := absolute(path)
	switch {
	case !outerKnown || !innerKnown:
		return false
	case inner == outer:
		return true
	default:
		return strings.HasPrefix(inner, endedWithASeparator(outer))
	}
}

func endedWithASeparator(root string) string {
	separator := string(os.PathSeparator)
	return strings.TrimSuffix(root, separator) + separator
}

func absolute(path string) (string, bool) {
	if path == "" {
		return "", false
	}
	named := withHomeExpanded(path)
	from, known := workingDirectory(named)
	return walked(from, strings.TrimPrefix(named, filepath.VolumeName(named))), known
}

func workingDirectory(path string) (string, bool) {
	here, err := os.Getwd()
	volume := filepath.VolumeName(path)
	switch {
	case filepath.IsAbs(path):
		return volume + string(os.PathSeparator), true
	case rooted(path):
		return filepath.VolumeName(here) + string(os.PathSeparator), err == nil
	case volume != "":
		return "", false
	default:
		return here, err == nil
	}
}

func rooted(path string) bool {
	return path != "" && os.IsPathSeparator(path[0])
}

func walked(here, path string) string {
	for _, name := range strings.Split(filepath.ToSlash(path), "/") {
		switch name {
		case "", ".":
		case "..":
			here = filepath.Dir(here)
		default:
			here = followed(filepath.Join(here, name))
		}
	}
	return here
}

func followed(path string) string {
	if leads, err := filepath.EvalSymlinks(path); err == nil {
		return leads
	}
	return path
}

func withHomeExpanded(path string) string {
	home, err := os.UserHomeDir()
	switch {
	case err != nil, !strings.HasPrefix(path, "~"):
		return path
	case path == "~":
		return home
	case strings.HasPrefix(path, "~/"):
		return home + path[1:]
	default:
		return path
	}
}
