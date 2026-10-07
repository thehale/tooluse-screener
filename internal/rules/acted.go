// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"slices"

	"github.com/thehale/tooluse-screener/internal/directories"
)

func directoriesActedIn(destinations, targets []string) []string {
	dirs := slices.Clone(targets)
	for _, here := range destinations {
		for _, there := range orHere(targets) {
			dirs = append(dirs, directories.Within(here, there))
		}
	}
	return dirs
}

func orHere(targets []string) []string {
	if len(targets) == 0 {
		return []string{"."}
	} else {
		return targets
	}
}
