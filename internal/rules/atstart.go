// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import commands "github.com/thehale/tooluse-screener/internal/command"

type AtStart struct {
	Expression Expression
}

func (a AtStart) Find(command commands.Command) (Match, bool) {
	bareCommand := command.WithoutAssignments()
	match, found := a.Expression.Find(bareCommand)
	if found && match.Start == 0 && isWordBoundary(bareCommand.Text, match.Width) {
		return Match{0, match.Width}, true
	} else {
		return Match{}, false
	}
}

func (a AtStart) String() string {
	return a.Expression.String()
}

func isWordBoundary(text string, at int) bool {
	return at > 0 && (at == len(text) || text[at] == ' ')
}
