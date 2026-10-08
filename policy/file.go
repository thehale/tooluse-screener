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
	deniedSide, err := f.Denied.side(rules.Rule{})
	if err != nil {
		return Policy{}, err
	}
	trusted := rules.TrustedGitDirectories(f.Trusted)
	allowedSide, err := f.Allowed.side(rules.Rule{AtStart: true, TrustedGitDirectories: &trusted})
	return Policy{denied: deniedSide, allowed: allowedSide}, err
}
