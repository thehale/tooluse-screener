// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules

func (g Group) AsAllowed(trusted []string) Group {
	allowed := make(Group, len(g))
	for index, rule := range g {
		rule.Expression = TrustedGitDirectories{Dirs: trusted, Expression: AtStart{Expression: rule.Expression}}
		allowed[index] = rule
	}
	return allowed
}
