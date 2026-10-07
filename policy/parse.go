// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/thehale/tooluse-screener/internal/rules"
)

func Parse(contents []byte) (Policy, error) {
	var configuration file
	if err := yaml.Unmarshal(contents, &configuration); err != nil {
		return Policy{}, err
	}
	return rulesScope(configuration)
}

func rulesScope(configuration file) (Policy, error) {
	deniedSide, err := sideFrom(configuration.Denied, denial)
	if err != nil {
		return Policy{}, err
	}
	allowedSide, err := sideFrom(configuration.Allowed, allowanceIn(configuration.Trusted))
	return Policy{denied: deniedSide, allowed: allowedSide}, err
}

func sideFrom(entries []entry, wrap func(rules.Rule) rules.Rule) (side, error) {
	commandRules, err := commandRulesOf(entries, wrap)
	if err != nil {
		return side{}, err
	}
	pathRules, err := pathRulesOf(entries)
	return side{commandRules: commandRules, pathRules: pathRules}, err
}

func pathRulesOf(entries []entry) (rules.Globs, error) {
	var pathRules rules.Globs
	for _, one := range entries {
		list, err := one.pathRules()
		if err != nil {
			return nil, err
		}
		pathRules = append(pathRules, list...)
	}
	return pathRules, nil
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
		glob, err := rules.NewGlob(text, e.Reason, e.Name)
		if err != nil {
			return nil, err
		}
		pathRules = append(pathRules, glob)
	}
	return pathRules, nil
}

func denial(rule rules.Rule) rules.Rule {
	return rule
}

func allowanceIn(trusted []string) func(rules.Rule) rules.Rule {
	return func(rule rules.Rule) rules.Rule {
		return rules.NewReaching(rules.NewOpening(rule), trusted)
	}
}

func commandRulesOf(entries []entry, wrap func(rules.Rule) rules.Rule) (rules.Group, error) {
	var commandRules []rules.Rule
	for _, one := range entries {
		list, err := one.commandRules(wrap)
		if err != nil {
			return rules.Group{}, err
		}
		commandRules = append(commandRules, list...)
	}
	return rules.Group(commandRules), nil
}

func (e entry) commandRules(wrap func(rules.Rule) rules.Rule) ([]rules.Rule, error) {
	commandRules, err := e.unscopedRules(wrap)
	if err != nil {
		return nil, err
	}
	commandRules, err = scopedRules(commandRules, e.Only)
	if err != nil {
		return nil, err
	}
	return addendedRules(commandRules, e.Addendum)
}

func scopedRules(commandRules []rules.Rule, only *scope) ([]rules.Rule, error) {
	if only == nil {
		return commandRules, nil
	} else {
		ruleScope, err := only.rulesScope()
		return rulesScopedTo(commandRules, ruleScope), err
	}
}

func rulesScopedTo(commandRules []rules.Rule, scope rules.Scope) []rules.Rule {
	rulesInScope := make([]rules.Rule, 0, len(commandRules))
	for _, rule := range commandRules {
		rulesInScope = append(rulesInScope, rules.NewScoped(rule, scope))
	}
	return rulesInScope
}

func addendedRules(commandRules []rules.Rule, additions []addition) ([]rules.Rule, error) {
	addenda, err := addendaOf(additions)
	rulesWithAddenda := make([]rules.Rule, 0, len(commandRules))
	for _, rule := range commandRules {
		rulesWithAddenda = append(rulesWithAddenda, rules.NewAddended(rule, addenda))
	}
	return rulesWithAddenda, err
}

func addendaOf(additions []addition) ([]rules.Addendum, error) {
	var addenda []rules.Addendum
	for _, one := range additions {
		if one.Only == nil || strings.TrimSpace(one.Reason) == "" {
			return nil, fmt.Errorf("an addendum is a mapping with `only` and `reason`: %+v", one)
		}
		scope, err := one.Only.rulesScope()
		if err != nil {
			return nil, err
		}
		addenda = append(addenda, rules.NewAddendum(scope, one.Reason))
	}
	return addenda, nil
}

func (e entry) unscopedRules(wrap func(rules.Rule) rules.Rule) ([]rules.Rule, error) {
	if len(e.Commands) == 0 && len(e.Patterns) == 0 && len(e.Paths) == 0 {
		return nil, fmt.Errorf("an entry is a command, or a mapping with `commands`, `patterns` or `paths`: %+v", e)
	} else {
		return e.writtenRules(wrap)
	}
}

func (e entry) writtenRules(wrap func(rules.Rule) rules.Rule) ([]rules.Rule, error) {
	var commandRules []rules.Rule
	for _, text := range e.Commands {
		commandRules = append(commandRules, wrap(e.commandRule(text)))
	}
	for _, expression := range e.Patterns {
		pattern, err := rules.NewPattern(expression, e.Reason, e.Name)
		if err != nil {
			return nil, err
		}
		commandRules = append(commandRules, wrap(pattern))
	}
	return commandRules, nil
}

func (e entry) commandRule(text string) rules.Rule {
	if invocation, isGit := gitInvocation(text); isGit {
		return rules.NewGitSubcommand(invocation[0], invocation[1:], e.Reason, e.Name)
	} else {
		return rules.NewWords(text, e.Reason, e.Name)
	}
}

func gitInvocation(text string) (arguments []string, isGit bool) {
	if words := strings.Fields(text); len(words) < 2 || words[0] != "git" {
		return nil, false
	} else {
		return words[1:], true
	}
}
