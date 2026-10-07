// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/thehale/tooluse-screener/internal/rules"
)

type file struct {
	Trusted []string `yaml:"trusted_git_directories"`
	Denied  []entry  `yaml:"denied"`
	Allowed []entry  `yaml:"allowed"`
}

type entry struct {
	Commands    texts      `yaml:"commands"`
	Patterns    texts      `yaml:"patterns"`
	Paths       texts      `yaml:"paths"`
	Reason      string     `yaml:"reason"`
	Name        string     `yaml:"name"`
	Description string     `yaml:"description"`
	Only        *scope     `yaml:"only"`
	Addendum    []addition `yaml:"addendum"`
}

type addition struct {
	Only   *scope `yaml:"only"`
	Reason string `yaml:"reason"`
}

type scope struct {
	Dirs     texts       `yaml:"dirs"`
	Branches []condition `yaml:"branches"`
}

type condition struct {
	Onto string `yaml:"-"`
	Not  texts  `yaml:"not"`
}

func (c *condition) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		return node.Decode(&c.Onto)
	}
	type mapping condition
	return node.Decode((*mapping)(c))
}

type texts []string

func (s scope) built() (rules.Scope, error) {
	branches, err := branchesNamed(s.Branches)
	return rules.NewScope(s.Dirs, branches), err
}

func branchesNamed(conditions []condition) (rules.Branches, error) {
	var named rules.Branches
	for _, written := range conditions {
		switch {
		case written.Onto != "":
			named.Onto = append(named.Onto, written.Onto)
		case len(written.Not) > 0:
			named.NotOnto = append(named.NotOnto, written.Not...)
		default:
			return rules.Branches{}, fmt.Errorf("a branch is a name, or a mapping with `not`: %+v", written)
		}
	}
	return named, nil
}

func (e *entry) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		return node.Decode(&e.Commands)
	}
	type mapping entry
	if err := node.Decode((*mapping)(e)); err != nil {
		return err
	}
	return renamed(e.Description)
}

func renamed(description string) error {
	switch description {
	case "":
		return nil
	default:
		return fmt.Errorf("`description` is now `name`: %s", description)
	}
}

func (t *texts) UnmarshalYAML(node *yaml.Node) error {
	var read []string
	for _, written := range writtenIn(node) {
		if !isText(written) {
			return fmt.Errorf("a command or expression is written as text, and this is not: %v", written.Value)
		}
		read = append(read, written.Value)
	}
	*t = read
	return nil
}

func writtenIn(node *yaml.Node) []*yaml.Node {
	if node.Kind == yaml.SequenceNode {
		return node.Content
	}
	return []*yaml.Node{node}
}

func isText(written *yaml.Node) bool {
	return written.Tag == "!!str" && strings.TrimSpace(written.Value) != ""
}
