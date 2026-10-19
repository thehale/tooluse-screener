// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	commands "github.com/thehale/tooluse-screener/internal/command"
	"github.com/thehale/tooluse-screener/internal/directories"
)

type TrustedGitDirectories struct {
	Dirs       []string
	workingDir directories.WorkingDir
}

func (d TrustedGitDirectories) InDir(workingDir directories.WorkingDir) TrustedGitDirectories {
	d.workingDir = workingDir
	return d
}

func (d TrustedGitDirectories) hasTargetsOf(command commands.Command) bool {
	dirs := d.workingDir.ResolveAll(command.DirectoriesActedIn(command.TargetDirectories()))
	return command.Git == nil || directories.AreAllUnder(dirs, d.workingDir.ResolveAll(d.Dirs))
}
