// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import commands "github.com/thehale/tooluse-screener/internal/command"

type Opening struct {
	Rule
}

func NewOpening(rule Rule) Rule {
	return Opening{rule}
}

func (o Opening) At(command commands.Command) int {
	if o.Span(command) > 0 {
		return 0
	} else {
		return -1
	}
}

func (o Opening) Span(command commands.Command) int {
	bareCommand := command.WithoutAssignments()
	span := o.Rule.Span(bareCommand)
	if o.Rule.At(bareCommand) == 0 && isWordBoundary(bareCommand.Text, span) {
		return span
	} else {
		return 0
	}
}

func isWordBoundary(text string, at int) bool {
	return at > 0 && (at == len(text) || text[at] == ' ')
}
