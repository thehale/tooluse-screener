// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/thehale/tooluse-screener/internal/rules"
)

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
	} else if oldName, hasDescription := description(node); hasDescription {
		return fmt.Errorf("`description` is now `name`: %s", oldName)
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

func (e entry) pathRules() (rules.Globs, error) {
	switch {
	case len(e.Paths) == 0:
		return nil, nil
	case len(e.Commands) > 0 || len(e.Patterns) > 0:
		return nil, fmt.Errorf("an entry names `paths` or commands, never both: %+v", e)
	case e.Only != nil || len(e.Addendum) > 0:
		return nil, fmt.Errorf("`only` and `addendum` scope commands, and an entry with `paths` has none to scope: %+v", e)
	default:
		return e.globs()
	}
}

func (e entry) globs() (rules.Globs, error) {
	var pathRules rules.Globs
	for _, text := range e.Paths {
		glob, err := rules.NewGlob(text, e.Wording)
		if err != nil {
			return nil, err
		}
		pathRules = append(pathRules, glob)
	}
	return pathRules, nil
}

func (e entry) commandRules(wrap func(rules.Rule) rules.Rule) (rules.Group, error) {
	rulesAsWritten, err := e.unscopedRules()
	commandRules := make(rules.Group, 0, len(rulesAsWritten))
	for _, rule := range rulesAsWritten {
		commandRules = append(commandRules, e.addendedRule(e.scopedRule(wrap(rule))))
	}
	return commandRules, err
}

func (e entry) scopedRule(rule rules.Rule) rules.Rule {
	if e.Only == nil {
		return rule
	} else {
		return rules.NewScoped(rule, e.Only.Scope)
	}
}

func (e entry) addendedRule(rule rules.Rule) rules.Rule {
	addenda := make([]rules.Addendum, 0, len(e.Addendum))
	for _, block := range e.Addendum {
		addenda = append(addenda, block.Addendum)
	}
	return rules.NewAddended(rule, addenda)
}

func (e entry) unscopedRules() (rules.Group, error) {
	if len(e.Commands) == 0 && len(e.Patterns) == 0 && len(e.Paths) == 0 {
		return nil, fmt.Errorf("an entry is a command, or a mapping with `commands`, `patterns` or `paths`: %+v", e)
	} else {
		return e.writtenRules()
	}
}

func (e entry) writtenRules() (rules.Group, error) {
	var commandRules rules.Group
	for _, text := range e.Commands {
		commandRules = append(commandRules, e.commandRule(text))
	}
	for _, expression := range e.Patterns {
		pattern, err := rules.NewPattern(expression, e.Wording)
		if err != nil {
			return nil, err
		}
		commandRules = append(commandRules, pattern)
	}
	return commandRules, nil
}

func (e entry) commandRule(text string) rules.Rule {
	if invocation, isGit := gitInvocation(text); isGit {
		return rules.NewGitSubcommand(invocation[0], invocation[1:], e.Wording)
	} else {
		return rules.NewWords(text, e.Wording)
	}
}

func gitInvocation(text string) (arguments []string, isGit bool) {
	if words := strings.Fields(text); len(words) < 2 || words[0] != "git" {
		return nil, false
	} else {
		return words[1:], true
	}
}
