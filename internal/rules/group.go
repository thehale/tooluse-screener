// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	commands "github.com/thehale/tooluse-screener/internal/command"
	"github.com/thehale/tooluse-screener/internal/lists"
)

type Group []Rule

func (g Group) HasMatchFor(command commands.Command) bool {
	return lists.Some(g, func(rule Rule) bool { return rule.At(command) >= 0 })
}

func (g Group) RulesMatching(command commands.Command) Group {
	var matches Group
	for _, rule := range g {
		if rule.At(command) >= 0 {
			matches = append(matches, rule)
		}
	}
	return matches
}

func (g Group) Span(command commands.Command) int {
	maxSpan := 0
	for _, rule := range g {
		maxSpan = max(maxSpan, rule.Span(command))
	}
	return maxSpan
}
