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

// Matching returns the rules inside this one that matched, not this one.
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
