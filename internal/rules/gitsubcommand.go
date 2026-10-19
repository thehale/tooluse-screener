// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"slices"
	"strings"

	commands "github.com/thehale/tooluse-screener/internal/command"
)

type GitSubcommand struct {
	subcommand    string
	requiredWords []string
}

func NewGitSubcommand(subcommand string, requiredWords []string) GitSubcommand {
	return GitSubcommand{subcommand, requiredWords}
}

func (g GitSubcommand) isMatch(command commands.Command) bool {
	return command.Git != nil && command.Git.IsA(g.subcommand) && command.Git.HasAll(g.requiredWords)
}

func (g GitSubcommand) Find(command commands.Command) (Match, bool) {
	if g.isMatch(command) {
		return Match{0, endOf(command.Text, append([]string{g.subcommand}, g.requiredWords...))}, true
	} else {
		return Match{}, false
	}
}

func endOf(command string, words []string) int {
	end, at := 0, 0
	for _, word := range strings.Fields(command) {
		at = strings.Index(command[at:], word) + at
		if slices.Contains(words, word) {
			end = max(end, at+len(word))
		}
		at += len(word)
	}
	return end
}

func (g GitSubcommand) String() string {
	return strings.Join(append([]string{"git", g.subcommand}, g.requiredWords...), " ")
}
