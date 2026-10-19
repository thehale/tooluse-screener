// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"github.com/thehale/tooluse-screener/internal/directories"
	"github.com/thehale/tooluse-screener/internal/lists"
)

type Globs []Glob

func (g Globs) InDir(workingDir directories.WorkingDir) Globs {
	return lists.Map(g, func(glob Glob) Glob { return glob.InDir(workingDir) })
}

func (g Globs) GlobsMatching(path string) Globs {
	var matches Globs
	for _, glob := range g {
		if glob.IsMatchFor(path) {
			matches = append(matches, glob)
		}
	}
	return matches
}
