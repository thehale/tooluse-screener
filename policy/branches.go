// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"cmp"

	"gopkg.in/yaml.v3"

	"github.com/thehale/tooluse-screener/internal/rules"
)

type branchList rules.Branches

func (b *branchList) UnmarshalYAML(node *yaml.Node) error {
	var items []yaml.Node
	if err := node.Decode(&items); err != nil {
		return err
	}
	for _, item := range items {
		if err := b.add(item); err != nil {
			return err
		}
	}
	return nil
}

func (b *branchList) add(item yaml.Node) error {
	var exclusion struct {
		Not texts `yaml:"not"`
	}
	if item.Kind == yaml.ScalarNode {
		b.Onto = append(b.Onto, item.Value)
		return nil
	} else if err := item.Decode(&exclusion); err != nil || len(exclusion.Not) == 0 {
		return cmp.Or(err, refusalAt(item.Line, "a branch is a name, or a mapping with `not`"))
	} else {
		b.NotOnto = append(b.NotOnto, exclusion.Not...)
		return nil
	}
}
