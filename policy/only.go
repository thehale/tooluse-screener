// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"gopkg.in/yaml.v3"

	"github.com/thehale/tooluse-screener/internal/rules"
)

type onlyBlock struct {
	rules.Scope
}

var onlyFields = fields{
	{"dirs", shape{yaml.ScalarNode | yaml.SequenceNode, "a directory, or a list of them"}},
	{"branches", shape{yaml.SequenceNode, "a list of branches"}},
}

func (o *onlyBlock) UnmarshalYAML(node *yaml.Node) error {
	var decoded struct {
		Dirs     texts      `yaml:"dirs"`
		Branches branchList `yaml:"branches"`
	}
	if err := onlyFields.check(node); err != nil {
		return err
	} else if err := node.Decode(&decoded); err != nil {
		return err
	} else {
		o.Scope = rules.Scope{Dirs: decoded.Dirs, Branches: rules.Branches(decoded.Branches)}
		return nil
	}
}
