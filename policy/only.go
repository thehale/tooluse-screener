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

func (o *onlyBlock) UnmarshalYAML(node *yaml.Node) error {
	var fields struct {
		Dirs     texts      `yaml:"dirs"`
		Branches branchList `yaml:"branches"`
	}
	err := node.Decode(&fields)
	o.Scope = rules.Scope{Dirs: fields.Dirs, Branches: rules.Branches(fields.Branches)}
	return err
}
