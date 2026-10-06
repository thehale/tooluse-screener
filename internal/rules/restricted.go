// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import commands "github.com/thehale/tooluse-screener/internal/command"

type Restricted struct {
	rule  Rule
	scope Scope
}

func NewRestricted(rule Rule, scope Scope) Rule {
	return Restricted{rule, scope}
}

func (r Restricted) Matches(command commands.Command) bool {
	return r.At(command) >= 0
}

func (r Restricted) At(command commands.Command) int {
	switch {
	case r.scope.includes(command):
		return r.rule.At(command)
	default:
		return -1
	}
}

func (r Restricted) Span(command commands.Command) int {
	switch {
	case r.scope.includes(command):
		return r.rule.Span(command)
	default:
		return 0
	}
}

func (r Restricted) Reason(command commands.Command) string {
	return r.rule.Reason(command)
}

func (r Restricted) String() string {
	return r.rule.String()
}
