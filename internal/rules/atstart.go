// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import commands "github.com/thehale/tooluse-screener/internal/command"

type AtStart struct {
	Expression Expression
}

func (a AtStart) At(command commands.Command) int {
	if a.Span(command) > 0 {
		return 0
	} else {
		return -1
	}
}

func (a AtStart) Span(command commands.Command) int {
	bareCommand := command.WithoutAssignments()
	span := a.Expression.Span(bareCommand)
	if a.Expression.At(bareCommand) == 0 && isWordBoundary(bareCommand.Text, span) {
		return span
	} else {
		return 0
	}
}

func (a AtStart) String() string {
	return a.Expression.String()
}

func isWordBoundary(text string, at int) bool {
	return at > 0 && (at == len(text) || text[at] == ' ')
}
