// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package directories

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/thehale/tooluse-screener/internal/lists"
)

func AreAllUnder(paths, roots []string) bool {
	return lists.Every(paths, func(path string) bool {
		return lists.Some(roots, func(root string) bool { return isInside(root, path) })
	})
}

func isInside(root, path string) bool {
	outer, isOuterKnown := absolutePath(root)
	inner, isInnerKnown := absolutePath(path)
	switch {
	case !isOuterKnown || !isInnerKnown:
		return false
	case inner == outer:
		return true
	default:
		return strings.HasPrefix(inner, withTrailingSeparator(outer))
	}
}

func withTrailingSeparator(root string) string {
	separator := string(os.PathSeparator)
	return strings.TrimSuffix(root, separator) + separator
}

func RealPath(path string) string {
	realPath, _ := absolutePath(path)
	return realPath
}

func absolutePath(path string) (string, bool) {
	if expandedPath := WithHomeExpanded(path); expandedPath == "" || strings.HasPrefix(expandedPath, "~") {
		return "", false
	} else {
		from, isKnown := workingDirectory(expandedPath)
		return endpoint(from, strings.TrimPrefix(expandedPath, filepath.VolumeName(expandedPath))), isKnown
	}
}

func workingDirectory(path string) (string, bool) {
	here, err := os.Getwd()
	volume := filepath.VolumeName(path)
	switch {
	case filepath.IsAbs(path):
		return volume + string(os.PathSeparator), true
	case isRooted(path):
		return filepath.VolumeName(here) + string(os.PathSeparator), err == nil
	case volume != "":
		return "", false
	default:
		return here, err == nil
	}
}

func isRooted(path string) bool {
	return path != "" && os.IsPathSeparator(path[0])
}

func endpoint(here, path string) string {
	for _, name := range strings.Split(filepath.ToSlash(path), "/") {
		switch name {
		case "", ".":
		case "..":
			here = filepath.Dir(here)
		default:
			here = symlinkTarget(filepath.Join(here, name))
		}
	}
	return here
}

func symlinkTarget(path string) string {
	if leads, err := filepath.EvalSymlinks(path); err == nil {
		return leads
	} else {
		return path
	}
}

func WithHomeExpanded(path string) string {
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
