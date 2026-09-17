// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

import "github.com/thehale/tooluse-screener/internal/lists"

type Group struct {
	described
	rules []Rule
	name  string
}

func NewGroup(name string, rules ...Rule) Group {
	return Group{described{}, rules, name}
}

func (g Group) Matches(command string) bool {
	return lists.Some(g.rules, func(rule Rule) bool { return rule.Matches(command) })
}

func (g Group) Matching(command string) []Rule {
	var matched []Rule
	for _, rule := range g.rules {
		matched = append(matched, matchingOne(rule, command)...)
	}
	return matched
}

func matchingOne(rule Rule, command string) []Rule {
	if nested, grouped := rule.(Group); grouped {
		return nested.Matching(command)
	}
	if rule.Matches(command) {
		return []Rule{rule}
	}
	return nil
}

func (g Group) String() string {
	return g.describes(g.name)
}

func (g Group) At(command string) int {
	earliest := -1
	for _, rule := range g.rules {
		if at := rule.At(command); at >= 0 && (earliest < 0 || at < earliest) {
			earliest = at
		}
	}
	return earliest
}

func (g Group) Span(command string) int {
	widest := 0
	for _, rule := range g.rules {
		widest = max(widest, rule.Span(command))
	}
	return widest
}
