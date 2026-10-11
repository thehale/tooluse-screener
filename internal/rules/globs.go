// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package rules

type Globs []Glob

func (g Globs) GlobsMatching(path string) Globs {
	var matches Globs
	for _, glob := range g {
		if glob.IsMatchFor(path) {
			matches = append(matches, glob)
		}
	}
	return matches
}
