// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	commands "github.com/thehale/tooluse-screener/internal/command"
	"github.com/thehale/tooluse-screener/internal/directories"
	"github.com/thehale/tooluse-screener/internal/git"
)

type Restricted struct {
	rule     Rule
	dirs     []string
	branches Branches
}

func NewRestricted(rule Rule, dirs []string, branches Branches) Rule {
	return Restricted{rule, dirs, branches}
}

func (r Restricted) Matches(command commands.Command) bool {
	return r.At(command) >= 0
}

func (r Restricted) At(command commands.Command) int {
	switch {
	case r.holds(command):
		return r.rule.At(command)
	default:
		return -1
	}
}

func (r Restricted) holds(command commands.Command) bool {
	invocation := git.Read(command.Text)
	acted := actedIn(command.Moved, orHere(pointedAt(invocation)))
	return r.runsInADir(acted) && r.landsOnABranch(invocation, acted)
}

func (r Restricted) runsInADir(acted []string) bool {
	return len(r.dirs) == 0 || directories.AllUnder(acted, r.dirs)
}

func (r Restricted) landsOnABranch(invocation git.Invocation, acted []string) bool {
	switch {
	case r.branches.unsaid():
		return true
	default:
		branch, known := git.Landing(invocation, acted)
		return known && r.branches.hold(branch)
	}
}

func (r Restricted) Span(command commands.Command) int {
	switch {
	case r.holds(command):
		return r.rule.Span(command)
	default:
		return 0
	}
}

func (r Restricted) Note() string {
	return r.rule.Note()
}

func (r Restricted) String() string {
	return r.rule.String()
}
