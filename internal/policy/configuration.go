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
	Commands    texts  `yaml:"commands"`
	Patterns    texts  `yaml:"patterns"`
	Reason      string `yaml:"reason"`
	Description string `yaml:"description"`
}

type texts []string

type reading struct {
	command    func(text, note, description string) rules.Rule
	expression func(expression, note, description string) (rules.Rule, error)
	git        func(subcommand string, having []string, note, description string) rules.Rule
}

func built(configuration file) (Policy, error) {
	denied, err := group("denied", configuration.Denied, refusals)
	if err != nil {
		return Policy{}, err
	}
	allowed, err := group("allowed", configuration.Allowed, permissions(configuration.Trusted))
	if err != nil {
		return Policy{}, err
	}
	return Policy{Denied: denied, Allowed: allowed}, nil
}

var refusals = reading{rules.NewSubstring, rules.NewPattern, rules.NewGitSubcommand}

func permissions(trusted []string) reading {
	return reading{
		command:    rules.NewPrefix,
		expression: rules.NewLeading,
		git: func(subcommand string, having []string, note, description string) rules.Rule {
			return rules.NewGitSubcommandWithin(subcommand, having, trusted, note, description)
		},
	}
}

func group(name string, entries []entry, read reading) (rules.Group, error) {
	var found []rules.Rule
	for _, written := range entries {
		made, err := entryRules(written, read)
		if err != nil {
			return rules.Group{}, err
		}
		found = append(found, made...)
	}
	return rules.NewGroup(name, found...), nil
}

func entryRules(written entry, read reading) ([]rules.Rule, error) {
	if len(written.Commands) == 0 && len(written.Patterns) == 0 {
		return nil, fmt.Errorf("an entry is a command, or a mapping with `commands` or `patterns`: %+v", written)
	}
	var found []rules.Rule
	for _, text := range written.Commands {
		found = append(found, command(text, written.Reason, written.Description, read))
	}
	for _, expression := range written.Patterns {
		made, err := read.expression(expression, written.Reason, written.Description)
		if err != nil {
			return nil, err
		}
		found = append(found, made)
	}
	return found, nil
}

func command(text, note, description string, read reading) rules.Rule {
	if invocation, named := gitInvocation(text); named {
		return read.git(invocation[0], invocation[1:], note, description)
	}
	return read.command(text, note, description)
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
