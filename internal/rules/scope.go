// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	commands "github.com/thehale/tooluse-screener/internal/command"
	"github.com/thehale/tooluse-screener/internal/directories"
	"github.com/thehale/tooluse-screener/internal/git"
)

type Scope struct {
	Dirs       []string
	Branches   Branches
	workingDir directories.WorkingDir
}

func (s Scope) InDir(workingDir directories.WorkingDir) Scope {
	s.workingDir = workingDir
	return s
}

func (s Scope) isInScope(command commands.Command) bool {
	dirs := s.workingDir.ResolveAll(command.DirectoriesActedIn(directories.OrHere(command.TargetDirectories())))
	return s.isInADir(dirs) && s.isOnABranch(command.Git, dirs)
}

func (s Scope) isInADir(dirs []string) bool {
	return len(s.Dirs) == 0 || directories.AreAllUnder(dirs, s.workingDir.ResolveAll(s.Dirs))
}

func (s Scope) isOnABranch(invocation *git.Invocation, dirs []string) bool {
	switch {
	case s.Branches.isUnsaid():
		return true
	case invocation == nil:
		return false
	default:
		branch, isKnown := invocation.Landing(dirs)
		return isKnown && s.Branches.isListed(branch)
	}
}
