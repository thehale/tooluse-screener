// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package git

import (
	"slices"
	"strings"
)

type globalOption struct {
	name  string
	value string
	width int
}

func globalOptionAt(words []string) globalOption {
	name, value, isJoined := strings.Cut(words[0], "=")
	switch {
	case isJoined:
		return globalOption{name, value, 1}
	case hasAValue(name) && len(words) > 1:
		return globalOption{name, words[1], 2}
	default:
		return globalOption{name, "", 1}
	}
}

func hasAValue(option string) bool {
	withValues := []string{"-C", "-c", "--git-dir", "--namespace", "--work-tree"}
	return slices.Contains(withValues, option)
}
