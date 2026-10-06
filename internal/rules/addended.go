// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"strings"

	commands "github.com/thehale/tooluse-screener/internal/command"
)

type Addendum struct {
	scope  Scope
	reason string
}

func NewAddendum(scope Scope, reason string) Addendum {
	return Addendum{scope, reason}
}

type Addended struct {
	Rule
	addenda []Addendum
}

func NewAddended(rule Rule, addenda []Addendum) Rule {
	return Addended{rule, addenda}
}

func (a Addended) Reason(command commands.Command) string {
	reasons := []string{a.Rule.Reason(command)}
	for _, addendum := range a.addenda {
		if addendum.scope.includes(command) {
			reasons = append(reasons, addendum.reason)
		}
	}
	return strings.TrimSpace(strings.Join(reasons, " "))
}
