// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	commands "github.com/thehale/tooluse-screener/internal/command"
	"github.com/thehale/tooluse-screener/internal/directories"
	"github.com/thehale/tooluse-screener/internal/git"
)

type TrustedGitDirectories []string

func (d TrustedGitDirectories) hasTargetsOf(command commands.Command) bool {
	invocation := git.Read(command.Text)
	return !invocation.Git || directories.AreAllUnder(command.DirectoriesActedIn(invocation.TargetDirectories()), []string(d))
}
