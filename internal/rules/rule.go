// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"strings"

	commands "github.com/thehale/tooluse-screener/internal/command"
)

type Rule struct {
	Name       string
	Reason     string
	Expression Expression
	Only       Scope
	Addendums  []Addendum
}

func (r Rule) At(command commands.Command) int {
	if r.Only.isInScope(command) {
		return r.Expression.At(command)
	} else {
		return -1
	}
}

func (r Rule) Span(command commands.Command) int {
	if r.Only.isInScope(command) {
		return r.Expression.Span(command)
	} else {
		return 0
	}
}

func (r Rule) ReasonFor(command commands.Command) string {
	reasons := []string{r.Reason}
	for _, addendum := range r.Addendums {
		if addendum.Only.isInScope(command) {
			reasons = append(reasons, addendum.Reason)
		}
	}
	return strings.TrimSpace(strings.Join(reasons, " "))
}

func (r Rule) String() string {
	return nameOr(r.Name, r.Expression.String())
}
