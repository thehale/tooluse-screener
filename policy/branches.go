// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"gopkg.in/yaml.v3"

	"github.com/thehale/tooluse-screener/internal/rules"
)

type branchList rules.Branches

var branchShape = shape{yaml.ScalarNode | yaml.MappingNode, "a branch name, or a mapping with `not` listing branches to exclude"}

var branchFields = fields{
	{"not", shape{yaml.ScalarNode | yaml.SequenceNode, "a branch name, or a list of them"}},
}

func (b *branchList) UnmarshalYAML(node *yaml.Node) error {
	var items []yaml.Node
	if err := node.Decode(&items); err != nil {
		return err
	}
	for _, item := range items {
		if err := b.add(&item); err != nil {
			return err
		}
	}
	return nil
}

func (b *branchList) add(item *yaml.Node) error {
	var exclusion struct {
		Not *texts `yaml:"not"`
	}
	if err := branchShape.check(item, "branch"); err != nil {
		return err
	} else if item.Kind == yaml.ScalarNode {
		b.Onto = append(b.Onto, item.Value)
		return nil
	} else if err := branchFields.check(item); err != nil {
		return err
	} else if err := item.Decode(&exclusion); err != nil {
		return err
	} else if exclusion.Not == nil {
		return refusalAt(item.Line, "branch missing `not`")
	} else if len(*exclusion.Not) == 0 {
		return refusalAt(item.Line, "`not` names no branch. List the branches to exclude.")
	} else {
		b.NotOnto = append(b.NotOnto, *exclusion.Not...)
		return nil
	}
}
