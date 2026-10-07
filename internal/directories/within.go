// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package directories

import (
	"path/filepath"
	"strings"
)

func Within(here, path string) string {
	switch {
	case here == "" || path == "":
		return ""
	case here == "." || isStandalone(path):
		return path
	default:
		return here + string(filepath.Separator) + path
	}
}

func isStandalone(path string) bool {
	return isRooted(path) || filepath.VolumeName(path) != "" || strings.HasPrefix(path, "~")
}

func Destination(here, path string) string {
	expandedPath := WithHomeExpanded(Within(here, path))
	if expandedPath == "" || strings.HasPrefix(expandedPath, "~") {
		return ""
	} else {
		return filepath.Clean(expandedPath)
	}
}
