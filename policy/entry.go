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
	Commands texts           `yaml:"commands"`
	Patterns texts           `yaml:"patterns"`
	Paths    texts           `yaml:"paths"`
	Name     string          `yaml:"name"`
	Reason   string          `yaml:"reason"`
	Only     *onlyBlock      `yaml:"only"`
	Addendum []addendumBlock `yaml:"addendum"`
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
		glob, err := rules.NewGlob(text)
		if err != nil {
			return nil, err
		}
		glob.Name, glob.Reason = e.Name, e.Reason
		pathRules = append(pathRules, glob)
	}
	return pathRules, nil
}

func (e entry) commandRules(base rules.Rule) (rules.Group, error) {
	expressions, err := e.expressions()
	commandRules := make(rules.Group, 0, len(expressions))
	for _, expression := range expressions {
		commandRule := base
		commandRule.Name = e.Name
		commandRule.Reason = e.Reason
		commandRule.Expression = expression
		commandRule.Only = e.scope()
		commandRule.Addendums = e.addendums()
		commandRules = append(commandRules, commandRule)
	}
	return commandRules, err
}

func (e entry) scope() rules.Scope {
	if e.Only == nil {
		return rules.Scope{}
	} else {
		return e.Only.Scope
	}
}

func (e entry) addendums() []rules.Addendum {
	addendums := make([]rules.Addendum, 0, len(e.Addendum))
	for _, block := range e.Addendum {
		addendums = append(addendums, block.Addendum)
	}
	return addendums
}

func (e entry) expressions() ([]rules.Expression, error) {
	if len(e.Commands) == 0 && len(e.Patterns) == 0 && len(e.Paths) == 0 {
		return nil, fmt.Errorf("an entry is a command, or a mapping with `commands`, `patterns` or `paths`: %+v", e)
	} else {
		return e.writtenExpressions()
	}
}

func (e entry) writtenExpressions() ([]rules.Expression, error) {
	var expressions []rules.Expression
	for _, text := range e.Commands {
		expressions = append(expressions, e.expression(text))
	}
	for _, text := range e.Patterns {
		pattern, err := rules.NewPattern(text)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, pattern)
	}
	return expressions, nil
}

func (e entry) expression(text string) rules.Expression {
	if invocation, isGit := gitInvocation(text); isGit {
		return rules.NewGitSubcommand(invocation[0], invocation[1:])
	} else {
		return rules.Words(text)
	}
}

func gitInvocation(text string) (arguments []string, isGit bool) {
	if words := strings.Fields(text); len(words) < 2 || words[0] != "git" {
		return nil, false
	} else {
		return words[1:], true
	}
}
