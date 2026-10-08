// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"slices"

	"github.com/thehale/tooluse-screener/internal/directories"
)

type Command struct {
	Text  string
	Moved []string
}

type Spanner interface {
	Span(command Command) int
}

func (c Command) IsAccountedForBy(rule Spanner) bool {
	words := c.WithoutAssignments().Text
	return words != "" && rule.Span(c) == len(words)
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
	return Command{Text: WithoutAssignments(c.Text), Moved: c.Moved}
}
