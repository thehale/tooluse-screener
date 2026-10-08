// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import "github.com/thehale/tooluse-screener/internal/rules"

type file struct {
	Trusted []string `yaml:"trusted_git_directories"`
	Denied  entries  `yaml:"denied"`
	Allowed entries  `yaml:"allowed"`
}

func (f file) policy() (Policy, error) {
	deniedSide, err := f.Denied.side(asWritten)
	if err != nil {
		return Policy{}, err
	}
	allowedSide, err := f.Allowed.side(allowRuleIn(f.Trusted))
	return Policy{denied: deniedSide, allowed: allowedSide}, err
}

func asWritten(rule rules.Rule) rules.Rule {
	return rule
}

func allowRuleIn(trustedDirs []string) func(rules.Rule) rules.Rule {
	return func(rule rules.Rule) rules.Rule {
		return rules.NewReaching(rules.NewOpening(rule), trustedDirs)
	}
}
