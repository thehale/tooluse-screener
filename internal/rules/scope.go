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

func (s Scope) isInScope(command commands.Command) bool {
	invocation := git.Read(command.Text)
	dirs := directoriesActedIn(command.Moved, orHere(invocation.TargetDirectories()))
	return s.isInADir(dirs) && s.isOnABranch(invocation, dirs)
}

func (s Scope) isInADir(dirs []string) bool {
	return len(s.dirs) == 0 || directories.AreAllUnder(dirs, s.dirs)
}

func (s Scope) isOnABranch(invocation git.Invocation, dirs []string) bool {
	if s.branches.isUnsaid() {
		return true
	} else {
		branch, isKnown := invocation.Landing(dirs)
		return isKnown && s.branches.isListed(branch)
	}
}
