// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import commands "github.com/thehale/tooluse-screener/internal/command"

type Scoped struct {
	Rule
	scope Scope
}

func NewScoped(rule Rule, scope Scope) Rule {
	return Scoped{rule, scope}
}

func (r Scoped) At(command commands.Command) int {
	if r.scope.isInScope(command) {
		return r.Rule.At(command)
	} else {
		return -1
	}
}

func (r Scoped) Span(command commands.Command) int {
	if r.scope.isInScope(command) {
		return r.Rule.Span(command)
	} else {
		return 0
	}
}
