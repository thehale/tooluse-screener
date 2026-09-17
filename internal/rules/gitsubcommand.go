// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"strings"

	"github.com/thehale/tooluse-screener/internal/directories"
	"github.com/thehale/tooluse-screener/internal/git"
)

type GitSubcommand struct {
	described
	subcommand string
	having     []string
	within     []string
	anywhere   bool
}

func NewGitSubcommand(subcommand string, having []string, note, description string) Rule {
	return GitSubcommand{described{note, description}, subcommand, having, nil, true}
}

func NewGitSubcommandWithin(subcommand string, having, roots []string, note, description string) Rule {
	return GitSubcommand{described{note, description}, subcommand, having, roots, false}
}

func (g GitSubcommand) Matches(command string) bool {
	invocation := git.Read(command)
	return invocation.IsA(g.subcommand) &&
		invocation.Carries(g.having) &&
		g.trusts(invocation.Directories)
}

func (g GitSubcommand) trusts(paths []string) bool {
	return g.anywhere || directories.AllUnder(paths, g.within)
}

func (g GitSubcommand) String() string {
	return g.describes(strings.Join(append([]string{"git", g.subcommand}, g.having...), " "))
}
