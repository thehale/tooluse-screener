// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

type Globs []Glob

func (g Globs) GlobsMatching(path string) Globs {
	var matched Globs
	for _, glob := range g {
		if glob.IsMatchFor(path) {
			matched = append(matched, glob)
		}
	}
	return matched
}
