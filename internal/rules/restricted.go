// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
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

func (r Restricted) Matches(command string) bool {
	return r.At(command) >= 0
}

func (r Restricted) At(command string) int {
	switch {
	case r.holds(command):
		return r.rule.At(command)
	default:
		return -1
	}
}

func (r Restricted) holds(command string) bool {
	invocation := git.Read(command)
	return r.runsInADir(invocation) && r.landsOnABranch(invocation)
}

func (r Restricted) runsInADir(invocation git.Invocation) bool {
	switch {
	case len(r.dirs) == 0:
		return true
	default:
		return directories.AllUnder(orHere(invocation.Directories), r.dirs)
	}
}

func orHere(pointed []string) []string {
	switch {
	case len(pointed) == 0:
		return []string{"."}
	default:
		return pointed
	}
}

func (r Restricted) landsOnABranch(invocation git.Invocation) bool {
	branch, known := git.Landing(invocation)
	switch {
	case r.branches.unsaid():
		return true
	default:
		return known && r.branches.hold(branch)
	}
}

func (r Restricted) Span(command string) int {
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
