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

func built(configuration file) (Policy, error) {
	denied, err := group("denied", configuration.Denied, asWritten)
	if err != nil {
		return Policy{}, err
	}
	allowed, err := group("allowed", configuration.Allowed, vouching(configuration.Trusted))
	if err != nil {
		return Policy{}, err
	}
	paths, err := pathsIn(configuration)
	if err != nil {
		return Policy{}, err
	}
	return Policy{Denied: denied, Allowed: allowed, Paths: paths}, nil
}

func pathsIn(configuration file) (Paths, error) {
	denied, err := globs(configuration.Denied)
	if err != nil {
		return Paths{}, err
	}
	allowed, err := globs(configuration.Allowed)
	return Paths{Denied: denied, Allowed: allowed}, err
}

func globs(entries []entry) ([]rules.Glob, error) {
	var found []rules.Glob
	for _, written := range entries {
		list, err := entryGlobs(written)
		if err != nil {
			return nil, err
		}
		found = append(found, list...)
	}
	return found, nil
}

func entryGlobs(written entry) ([]rules.Glob, error) {
	switch {
	case len(written.Paths) == 0:
		return nil, nil
	case len(written.Commands) > 0 || len(written.Patterns) > 0:
		return nil, fmt.Errorf("an entry names `paths` or commands, never both: %+v", written)
	case written.Only != nil || len(written.Addendum) > 0:
		return nil, fmt.Errorf("`only` and `addendum` scope commands, and an entry with `paths` has none to scope: %+v", written)
	}
	var found []rules.Glob
	for _, text := range written.Paths {
		made, err := rules.NewGlob(text, written.Reason, written.Name)
		if err != nil {
			return nil, err
		}
		found = append(found, made)
	}
	return found, nil
}

func asWritten(rule rules.Rule) rules.Rule {
	return rule
}

func vouching(trusted []string) func(rules.Rule) rules.Rule {
	return func(rule rules.Rule) rules.Rule {
		return rules.NewReaching(rules.NewOpening(rule), trusted)
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
	if err != nil {
		return nil, err
	}
	found, err = scoped(found, written.Only)
	if err != nil {
		return nil, err
	}
	return addended(found, written.Addendum)
}

func scoped(found []rules.Rule, only *scope) ([]rules.Rule, error) {
	switch only {
	case nil:
		return found, nil
	default:
		built, err := only.built()
		return within(found, built), err
	}
}

func within(found []rules.Rule, scope rules.Scope) []rules.Rule {
	narrower := make([]rules.Rule, 0, len(found))
	for _, rule := range found {
		narrower = append(narrower, rules.NewScoped(rule, scope))
	}
	return narrower
}

func (s scope) built() (rules.Scope, error) {
	branches, err := branchesNamed(s.Branches)
	return rules.NewScope(s.Dirs, branches), err
}

func addended(found []rules.Rule, written []addition) ([]rules.Rule, error) {
	addenda, err := addendaOf(written)
	reasoned := make([]rules.Rule, 0, len(found))
	for _, rule := range found {
		reasoned = append(reasoned, rules.NewAddended(rule, addenda))
	}
	return reasoned, err
}

func addendaOf(written []addition) ([]rules.Addendum, error) {
	var addenda []rules.Addendum
	for _, one := range written {
		if one.Only == nil || strings.TrimSpace(one.Reason) == "" {
			return nil, fmt.Errorf("an addendum is a mapping with `only` and `reason`: %+v", one)
		}
		scope, err := one.Only.built()
		if err != nil {
			return nil, err
		}
		addenda = append(addenda, rules.NewAddendum(scope, one.Reason))
	}
	return addenda, nil
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
	if len(written.Commands) == 0 && len(written.Patterns) == 0 && len(written.Paths) == 0 {
		return nil, fmt.Errorf("an entry is a command, or a mapping with `commands`, `patterns` or `paths`: %+v", written)
	}
	var found []rules.Rule
	for _, text := range written.Commands {
		found = append(found, read(command(text, written.Reason, written.Name)))
	}
	for _, expression := range written.Patterns {
		made, err := rules.NewPattern(expression, written.Reason, written.Name)
		if err != nil {
			return nil, err
		}
		found = append(found, read(made))
	}
	return found, nil
}

func command(text, reason, name string) rules.Rule {
	if invocation, named := gitInvocation(text); named {
		return rules.NewGitSubcommand(invocation[0], invocation[1:], reason, name)
	}
	return rules.NewWords(text, reason, name)
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
