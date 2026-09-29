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
	case here == "." || standsAlone(path):
		return path
	default:
		return here + string(filepath.Separator) + path
	}
}

func standsAlone(path string) bool {
	return rooted(path) || filepath.VolumeName(path) != "" || strings.HasPrefix(path, "~")
}
