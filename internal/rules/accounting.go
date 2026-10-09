// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import commands "github.com/thehale/tooluse-screener/internal/command"

func (r Rule) AccountsFor(command commands.Command) bool {
	return Group{r}.AccountsFor(command)
}

func (g Group) AccountsFor(command commands.Command) bool {
	words := command.WithoutAssignments().Text
	return words != "" && g.Span(command) == len(words)
}
