// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	commands "github.com/thehale/tooluse-screener/internal/command"
	"github.com/thehale/tooluse-screener/internal/directories"
	"github.com/thehale/tooluse-screener/internal/git"
)

type Reaching struct {
	rule    Rule
	trusted []string
}

func NewReaching(rule Rule, trusted []string) Rule {
	return Reaching{rule, trusted}
}

func (r Reaching) Matches(command commands.Command) bool {
	return r.At(command) >= 0
}

func (r Reaching) At(command commands.Command) int {
	switch {
	case r.reaches(command):
		return r.rule.At(command)
	default:
		return -1
	}
}

func (r Reaching) Span(command commands.Command) int {
	switch {
	case r.reaches(command):
		return r.rule.Span(command)
	default:
		return 0
	}
}

func (r Reaching) reaches(command commands.Command) bool {
	invocation := git.Read(command.Text)
	return !invocation.Git || directories.AllUnder(actedIn(command.Moved, invocation.PointedAt()), r.trusted)
}

func (r Reaching) Reason(command commands.Command) string {
	return r.rule.Reason(command)
}

func (r Reaching) String() string {
	return r.rule.String()
}
