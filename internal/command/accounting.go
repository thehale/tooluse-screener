// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command

type Spanner interface {
	Span(command Command) int
}

func (c Command) IsAccountedForBy(rule Spanner) bool {
	spoken := Spoken(c.Text)
	return spoken != "" && rule.Span(c) == len(spoken)
}
