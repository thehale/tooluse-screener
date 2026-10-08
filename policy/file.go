// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"gopkg.in/yaml.v3"

	"github.com/thehale/tooluse-screener/internal/rules"
)

type file struct {
	Trusted texts   `yaml:"trusted_git_directories"`
	Denied  entries `yaml:"denied"`
	Allowed entries `yaml:"allowed"`
}

var fileShape = shape{yaml.MappingNode, "a mapping with `denied`, `allowed` or `trusted_git_directories`"}

var fileFields = fields{
	{"trusted_git_directories", shape{yaml.SequenceNode, "a list of directories"}},
	{"denied", shape{yaml.SequenceNode, "a list of rules"}},
	{"allowed", shape{yaml.SequenceNode, "a list of rules"}},
}

func (f *file) UnmarshalYAML(node *yaml.Node) error {
	type mapping file
	if err := fileShape.check(node, "policy"); err != nil {
		return err
	} else if err := fileFields.check(node); err != nil {
		return err
	} else {
		return node.Decode((*mapping)(f))
	}
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
