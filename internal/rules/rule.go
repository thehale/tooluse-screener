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

func (r Rule) Find(command commands.Command) (Match, bool) {
	if r.Only.isInScope(command) {
		return r.Expression.Find(command)
	} else {
		return Match{}, false
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
