// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"cmp"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/thehale/tooluse-screener/internal/rules"
)

type file struct {
	Trusted []string `yaml:"trusted_git_directories"`
	Denied  entries  `yaml:"denied"`
	Allowed entries  `yaml:"allowed"`
}

type entries []entry

type entry struct {
	Commands      texts `yaml:"commands"`
	Patterns      texts `yaml:"patterns"`
	Paths         texts `yaml:"paths"`
	rules.Wording `yaml:",inline"`
	Only          *onlyBlock      `yaml:"only"`
	Addendum      []addendumBlock `yaml:"addendum"`
}

func (e *entry) UnmarshalYAML(node *yaml.Node) error {
	type mapping entry
	if node.Kind == yaml.ScalarNode {
		return node.Decode(&e.Commands)
	} else if retired, isWritten := description(node); isWritten {
		return fmt.Errorf("`description` is now `name`: %s", retired)
	} else {
		return node.Decode((*mapping)(e))
	}
}

func description(node *yaml.Node) (string, bool) {
	for index := 0; index+1 < len(node.Content); index += 2 {
		if node.Content[index].Value == "description" {
			return node.Content[index+1].Value, true
		}
	}
	return "", false
}

type onlyBlock struct {
	rules.Scope
}

func (o *onlyBlock) UnmarshalYAML(node *yaml.Node) error {
	var written struct {
		Dirs     texts      `yaml:"dirs"`
		Branches branchList `yaml:"branches"`
	}
	err := node.Decode(&written)
	o.Scope = rules.NewScope(written.Dirs, rules.Branches(written.Branches))
	return err
}

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
	var excluded struct {
		Not texts `yaml:"not"`
	}
	if item.Kind == yaml.ScalarNode {
		b.Onto = append(b.Onto, item.Value)
		return nil
	} else if err := item.Decode(&excluded); err != nil || len(excluded.Not) == 0 {
		return cmp.Or(err, fmt.Errorf("a branch is a name, or a mapping with `not`: %v", item.Value))
	} else {
		b.NotOnto = append(b.NotOnto, excluded.Not...)
		return nil
	}
}

type addendumBlock struct {
	rules.Addendum
}

func (a *addendumBlock) UnmarshalYAML(node *yaml.Node) error {
	var written struct {
		Only   *onlyBlock `yaml:"only"`
		Reason string     `yaml:"reason"`
	}
	if err := node.Decode(&written); err != nil {
		return err
	} else if written.Only == nil || strings.TrimSpace(written.Reason) == "" {
		return fmt.Errorf("an addendum is a mapping with `only` and `reason`: %+v", written)
	} else {
		a.Addendum = rules.NewAddendum(written.Only.Scope, written.Reason)
		return nil
	}
}

type texts []string

func (t *texts) UnmarshalYAML(node *yaml.Node) error {
	var values []string
	for _, item := range nodesIn(node) {
		if !isText(item) {
			return fmt.Errorf("a command or expression is written as text, and this is not: %v", item.Value)
		}
		values = append(values, item.Value)
	}
	*t = values
	return nil
}

func nodesIn(node *yaml.Node) []*yaml.Node {
	if node.Kind == yaml.SequenceNode {
		return node.Content
	} else {
		return []*yaml.Node{node}
	}
}

func isText(node *yaml.Node) bool {
	return node.Tag == "!!str" && strings.TrimSpace(node.Value) != ""
}
