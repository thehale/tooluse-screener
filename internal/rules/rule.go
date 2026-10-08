// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"strings"

	commands "github.com/thehale/tooluse-screener/internal/command"
)

type Rule struct {
	Name                  string
	Reason                string
	Expression            Expression
	Only                  Scope
	Addendums             []Addendum
	AtStart               bool
	TrustedGitDirectories *TrustedGitDirectories
}

func (r Rule) At(command commands.Command) int {
	if !r.isApplicableTo(command) {
		return -1
	} else if r.AtStart {
		return r.positionAtStart(command)
	} else {
		return r.Expression.At(command)
	}
}

func (r Rule) Span(command commands.Command) int {
	if !r.isApplicableTo(command) {
		return 0
	} else if r.AtStart {
		return r.spanAtStart(command)
	} else {
		return r.Expression.Span(command)
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

func (r Rule) isApplicableTo(command commands.Command) bool {
	return r.Only.isInScope(command) && (r.TrustedGitDirectories == nil || r.TrustedGitDirectories.hasTargetsOf(command))
}

func (r Rule) positionAtStart(command commands.Command) int {
	if r.spanAtStart(command) > 0 {
		return 0
	} else {
		return -1
	}
}

func (r Rule) spanAtStart(command commands.Command) int {
	bareCommand := command.WithoutAssignments()
	span := r.Expression.Span(bareCommand)
	if r.Expression.At(bareCommand) == 0 && isWordBoundary(bareCommand.Text, span) {
		return span
	} else {
		return 0
	}
}

func isWordBoundary(text string, at int) bool {
	return at > 0 && (at == len(text) || text[at] == ' ')
}
