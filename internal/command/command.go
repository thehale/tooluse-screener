// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"slices"

	"github.com/thehale/tooluse-screener/internal/directories"
)

type Command struct {
	Text   string
	Moved  []string
	Writes []string
	source string
}

func (c Command) DirectoriesActedIn(targets []string) []string {
	dirs := slices.Clone(targets)
	for _, here := range c.Moved {
		for _, there := range directories.OrHere(targets) {
			dirs = append(dirs, directories.Within(here, there))
		}
	}
	return dirs
}

func (c Command) WithoutAssignments() Command {
	return Command{Text: WithoutAssignments(c.Text), Moved: c.Moved, Writes: c.Writes}
}

func (c Command) PathsWritten() []string {
	var paths []string
	for _, target := range c.Writes {
		paths = append(paths, target)
		if !directories.IsStandalone(target) {
			for _, here := range c.Moved {
				paths = append(paths, directories.Within(here, target))
			}
		}
	}
	return paths
}
