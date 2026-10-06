// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"strings"

	commands "github.com/thehale/tooluse-screener/internal/command"
)

type Addendum struct {
	scope Scope
	note  string
}

func NewAddendum(scope Scope, note string) Addendum {
	return Addendum{scope, note}
}

type Addended struct {
	Rule
	addenda []Addendum
}

func NewAddended(rule Rule, addenda []Addendum) Rule {
	return Addended{rule, addenda}
}

func (a Addended) Note(command commands.Command) string {
	notes := []string{a.Rule.Note(command)}
	for _, addendum := range a.addenda {
		if addendum.scope.includes(command) {
			notes = append(notes, addendum.note)
		}
	}
	return strings.TrimSpace(strings.Join(notes, " "))
}
