// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	commands "github.com/thehale/tooluse-screener/internal/command"
	"github.com/thehale/tooluse-screener/internal/directories"
	"github.com/thehale/tooluse-screener/internal/git"
)

type Scope struct {
	Dirs     []string
	Branches Branches
}

func (s Scope) isInScope(command commands.Command) bool {
	invocation := git.Read(command.Text)
	dirs := command.DirectoriesActedIn(directories.OrHere(invocation.TargetDirectories()))
	return s.isInADir(dirs) && s.isOnABranch(invocation, dirs)
}

func (s Scope) isInADir(dirs []string) bool {
	return len(s.Dirs) == 0 || directories.AreAllUnder(dirs, s.Dirs)
}

func (s Scope) isOnABranch(invocation git.Invocation, dirs []string) bool {
	if s.Branches.isUnsaid() {
		return true
	} else {
		branch, isKnown := invocation.Landing(dirs)
		return isKnown && s.Branches.isListed(branch)
	}
}
