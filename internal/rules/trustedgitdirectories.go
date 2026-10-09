// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	commands "github.com/thehale/tooluse-screener/internal/command"
	"github.com/thehale/tooluse-screener/internal/directories"
	"github.com/thehale/tooluse-screener/internal/git"
)

type TrustedGitDirectories struct {
	Dirs       []string
	Expression Expression
}

func (d TrustedGitDirectories) At(command commands.Command) int {
	if d.hasTargetsOf(command) {
		return d.Expression.At(command)
	} else {
		return -1
	}
}

func (d TrustedGitDirectories) Span(command commands.Command) int {
	if d.hasTargetsOf(command) {
		return d.Expression.Span(command)
	} else {
		return 0
	}
}

func (d TrustedGitDirectories) String() string {
	return d.Expression.String()
}

func (d TrustedGitDirectories) hasTargetsOf(command commands.Command) bool {
	invocation := git.Read(command.Text)
	return !invocation.Git || directories.AreAllUnder(command.DirectoriesActedIn(invocation.TargetDirectories()), d.Dirs)
}
