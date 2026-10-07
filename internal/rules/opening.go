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
	switch {
	case o.Span(command) > 0:
		return 0
	default:
		return -1
	}
}

func (o Opening) Span(command commands.Command) int {
	spoken := commands.Command{Text: commands.Spoken(command.Text), Moved: command.Moved}
	accounted := o.Rule.Span(spoken)
	switch {
	case o.Rule.At(spoken) == 0 && endsAWord(spoken.Text, accounted):
		return accounted
	default:
		return 0
	}
}

func endsAWord(spoken string, at int) bool {
	return at > 0 && (at == len(spoken) || spoken[at] == ' ')
}
