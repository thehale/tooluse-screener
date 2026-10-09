// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import commands "github.com/thehale/tooluse-screener/internal/command"

type spanner interface {
	Span(command commands.Command) int
}

func (r Rule) AccountsFor(command commands.Command) bool {
	return accountsFor(r, command)
}

func (g Group) AccountsFor(command commands.Command) bool {
	return accountsFor(g, command)
}

func accountsFor(s spanner, command commands.Command) bool {
	words := command.WithoutAssignments().Text
	return words != "" && s.Span(command) == len(words)
}
