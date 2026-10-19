// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package rules

func (g Group) AsAllowed(trusted []string) Group {
	allowed := make(Group, len(g))
	for index, rule := range g {
		rule.Expression = AtStart{Expression: rule.Expression}
		rule.Trusted = &TrustedGitDirectories{Dirs: trusted}
		allowed[index] = rule
	}
	return allowed
}
