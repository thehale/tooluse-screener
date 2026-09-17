// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import commands "github.com/thehale/tooluse-screener/internal/command"

type Opening struct {
	rule Rule
}

func NewOpening(rule Rule) Rule {
	return Opening{rule}
}

func (o Opening) Matches(command string) bool {
	return o.At(command) >= 0
}

func (o Opening) At(command string) int {
	switch {
	case o.Span(command) > 0:
		return 0
	default:
		return -1
	}
}

func (o Opening) Span(command string) int {
	spoken := commands.Spoken(command)
	accounted := o.rule.Span(spoken)
	switch {
	case o.rule.At(spoken) == 0 && endsAWord(spoken, accounted):
		return accounted
	default:
		return 0
	}
}

func endsAWord(spoken string, at int) bool {
	return at > 0 && (at == len(spoken) || spoken[at] == ' ')
}

func (o Opening) Note() string {
	return o.rule.Note()
}

func (o Opening) String() string {
	return o.rule.String()
}
