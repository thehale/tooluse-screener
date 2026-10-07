// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	commands "github.com/thehale/tooluse-screener/internal/command"
	"github.com/thehale/tooluse-screener/internal/directories"
	"github.com/thehale/tooluse-screener/internal/git"
)

type Reaching struct {
	Rule
	trusted []string
}

func NewReaching(rule Rule, trusted []string) Rule {
	return Reaching{rule, trusted}
}

func (r Reaching) At(command commands.Command) int {
	if r.isReachable(command) {
		return r.Rule.At(command)
	} else {
		return -1
	}
}

func (r Reaching) Span(command commands.Command) int {
	if r.isReachable(command) {
		return r.Rule.Span(command)
	} else {
		return 0
	}
}

func (r Reaching) isReachable(command commands.Command) bool {
	invocation := git.Read(command.Text)
	return !invocation.Git || directories.AllUnder(directoriesActedIn(command.Moved, invocation.TargetDirectories()), r.trusted)
}
