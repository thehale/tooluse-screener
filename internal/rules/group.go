// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	commands "github.com/thehale/tooluse-screener/internal/command"
	"github.com/thehale/tooluse-screener/internal/lists"
)

type Group []Rule

func (g Group) Matches(command commands.Command) bool {
	return lists.Some(g, func(rule Rule) bool { return rule.At(command) >= 0 })
}

func (g Group) Matching(command commands.Command) Group {
	var matched Group
	for _, rule := range g {
		if rule.At(command) >= 0 {
			matched = append(matched, rule)
		}
	}
	return matched
}

func (g Group) Span(command commands.Command) int {
	widest := 0
	for _, rule := range g {
		widest = max(widest, rule.Span(command))
	}
	return widest
}
