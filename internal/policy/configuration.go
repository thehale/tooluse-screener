// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/thehale/tooluse-screener/internal/rules"
)

func Read(path string) (Policy, error) {
	written, err := os.ReadFile(path)
	if err != nil {
		return Policy{}, err
	}
	return Parse(written)
}

func Parse(written []byte) (Policy, error) {
	var configuration file
	if err := yaml.Unmarshal(written, &configuration); err != nil {
		return Policy{}, err
	}
	return built(configuration)
}

type file struct {
	Trusted []string `yaml:"trusted_git_directories"`
	Denied  []entry  `yaml:"denied"`
	Allowed []entry  `yaml:"allowed"`
}

type entry struct {
	Commands    texts        `yaml:"commands"`
	Patterns    texts        `yaml:"patterns"`
	Reason      string       `yaml:"reason"`
	Description string       `yaml:"description"`
	Only        *restriction `yaml:"only"`
}

type restriction struct {
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

func built(configuration file) (Policy, error) {
	denied, err := group("denied", configuration.Denied, asWritten)
	if err != nil {
		return Policy{}, err
	}
	allowed, err := group("allowed", configuration.Allowed, vouching(configuration.Trusted))
	if err != nil {
		return Policy{}, err
	}
	return Policy{Denied: denied, Allowed: allowed}, nil
}

func asWritten(rule rules.Rule) rules.Rule {
	return rule
}

func vouching(trusted []string) func(rules.Rule) rules.Rule {
	return func(rule rules.Rule) rules.Rule {
		return rules.NewOpening(rules.NewReaching(rule, trusted))
	}
}

func group(name string, entries []entry, read func(rules.Rule) rules.Rule) (rules.Group, error) {
	var found []rules.Rule
	for _, written := range entries {
		list, err := entryRules(written, read)
		if err != nil {
			return rules.Group{}, err
		}
		found = append(found, list...)
	}
	return rules.NewGroup(name, found...), nil
}

func entryRules(written entry, read func(rules.Rule) rules.Rule) ([]rules.Rule, error) {
	found, err := allRules(written, read)
	switch {
	case err != nil:
		return nil, err
	case written.Only == nil:
		return found, nil
	default:
		return restricted(found, *written.Only)
	}
}

func restricted(found []rules.Rule, only restriction) ([]rules.Rule, error) {
	branches, err := branchesNamed(only.Branches)
	if err != nil {
		return nil, err
	}
	narrower := make([]rules.Rule, 0, len(found))
	for _, rule := range found {
		narrower = append(narrower, rules.NewRestricted(rule, only.Dirs, branches))
	}
	return narrower, nil
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

func allRules(written entry, read func(rules.Rule) rules.Rule) ([]rules.Rule, error) {
	if len(written.Commands) == 0 && len(written.Patterns) == 0 {
		return nil, fmt.Errorf("an entry is a command, or a mapping with `commands` or `patterns`: %+v", written)
	}
	var found []rules.Rule
	for _, text := range written.Commands {
		found = append(found, read(command(text, written.Reason, written.Description)))
	}
	for _, expression := range written.Patterns {
		made, err := rules.NewPattern(expression, written.Reason, written.Description)
		if err != nil {
			return nil, err
		}
		found = append(found, read(made))
	}
	return found, nil
}

func command(text, note, description string) rules.Rule {
	if invocation, named := gitInvocation(text); named {
		return rules.NewGitSubcommand(invocation[0], invocation[1:], note, description)
	}
	return rules.NewWords(text, note, description)
}

func gitInvocation(text string) (spoken []string, named bool) {
	words := strings.Fields(text)
	if len(words) < 2 || words[0] != "git" {
		return nil, false
	}
	return words[1:], true
}

func (e *entry) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		return node.Decode(&e.Commands)
	}
	type mapping entry
	return node.Decode((*mapping)(e))
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
