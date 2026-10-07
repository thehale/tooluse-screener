// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"slices"
	"strings"

	commands "github.com/thehale/tooluse-screener/internal/command"
	"github.com/thehale/tooluse-screener/internal/git"
)

type GitSubcommand struct {
	wording
	subcommand string
	having     []string
}

func NewGitSubcommand(subcommand string, having []string, reason, name string) Rule {
	return GitSubcommand{wording{reason, name}, subcommand, having}
}

func (g GitSubcommand) isMatch(command commands.Command) bool {
	invocation := git.Read(command.Text)
	return invocation.IsA(g.subcommand) && invocation.HasAll(g.having)
}

func (g GitSubcommand) At(command commands.Command) int {
	if g.isMatch(command) {
		return 0
	} else {
		return -1
	}
}

func (g GitSubcommand) Span(command commands.Command) int {
	if g.isMatch(command) {
		return endOf(command.Text, append([]string{g.subcommand}, g.having...))
	} else {
		return 0
	}
}

func endOf(command string, wanted []string) int {
	end, at := 0, 0
	for _, word := range strings.Fields(command) {
		at = strings.Index(command[at:], word) + at
		if slices.Contains(wanted, word) {
			end = max(end, at+len(word))
		}
		at += len(word)
	}
	return end
}

func (g GitSubcommand) String() string {
	return g.nameOr(strings.Join(append([]string{"git", g.subcommand}, g.having...), " "))
}
