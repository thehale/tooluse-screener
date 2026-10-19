// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	commands "github.com/thehale/tooluse-screener/internal/command"
	"github.com/thehale/tooluse-screener/internal/directories"
	"github.com/thehale/tooluse-screener/internal/lists"
)

type Group []Rule

func (g Group) InDir(workingDir directories.WorkingDir) Group {
	return lists.Map(g, func(rule Rule) Rule { return rule.InDir(workingDir) })
}

func (g Group) RulesMatching(command commands.Command) Group {
	var matches Group
	for _, rule := range g {
		if _, found := rule.Find(command); found {
			matches = append(matches, rule)
		}
	}
	return matches
}

func (g Group) Span(command commands.Command) int {
	maxSpan := 0
	for _, rule := range g {
		if match, found := rule.Find(command); found {
			maxSpan = max(maxSpan, match.Width)
		}
	}
	return maxSpan
}
