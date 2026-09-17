// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"slices"
	"strings"

	"github.com/thehale/tooluse-screener/internal/git"
)

type GitSubcommand struct {
	described
	subcommand string
	having     []string
}

func NewGitSubcommand(subcommand string, having []string, note, description string) Rule {
	return GitSubcommand{described{note, description}, subcommand, having}
}

func (g GitSubcommand) Matches(command string) bool {
	invocation := git.Read(command)
	return invocation.IsA(g.subcommand) && invocation.Carries(g.having)
}

func (g GitSubcommand) At(command string) int {
	switch {
	case g.Matches(command):
		return 0
	default:
		return -1
	}
}

func (g GitSubcommand) Span(command string) int {
	switch {
	case g.Matches(command):
		return past(command, append([]string{g.subcommand}, g.having...))
	default:
		return 0
	}
}

func past(command string, wanted []string) int {
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

func (g GitSubcommand) required() string {
	return strings.Join(append([]string{"git", g.subcommand}, g.having...), " ")
}

func (g GitSubcommand) String() string {
	return g.describes(g.required())
}
