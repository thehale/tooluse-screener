// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command

type Spanner interface {
	Span(command Command) int
}

func (c Command) IsAccountedForBy(rule Spanner) bool {
	words := WithoutAssignments(c.Text)
	return words != "" && rule.Span(c) == len(words)
}
