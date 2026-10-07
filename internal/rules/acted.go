// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"slices"

	"github.com/thehale/tooluse-screener/internal/directories"
)

func actedIn(moved, pointed []string) []string {
	acted := slices.Clone(pointed)
	for _, here := range moved {
		for _, there := range orHere(pointed) {
			acted = append(acted, directories.Within(here, there))
		}
	}
	return acted
}

func orHere(pointed []string) []string {
	switch {
	case len(pointed) == 0:
		return []string{"."}
	default:
		return pointed
	}
}
