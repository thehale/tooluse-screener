// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	commands "github.com/thehale/tooluse-screener/internal/command"
	"github.com/thehale/tooluse-screener/internal/directories"
	"github.com/thehale/tooluse-screener/internal/git"
)

type Scope struct {
	dirs     []string
	branches Branches
}

func NewScope(dirs []string, branches Branches) Scope {
	return Scope{dirs, branches}
}

func (s Scope) includes(command commands.Command) bool {
	invocation := git.Read(command.Text)
	acted := actedIn(command.Moved, orHere(pointedAt(invocation)))
	return s.runsInADir(acted) && s.landsOnABranch(invocation, acted)
}

func (s Scope) runsInADir(acted []string) bool {
	return len(s.dirs) == 0 || directories.AllUnder(acted, s.dirs)
}

func (s Scope) landsOnABranch(invocation git.Invocation, acted []string) bool {
	switch {
	case s.branches.unsaid():
		return true
	default:
		branch, known := git.Landing(invocation, acted)
		return known && s.branches.hold(branch)
	}
}
