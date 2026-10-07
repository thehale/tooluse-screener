// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/thehale/tooluse-screener/internal/rules"
)

func Parse(written []byte) (Policy, error) {
	var configuration file
	if err := yaml.Unmarshal(written, &configuration); err != nil {
		return Policy{}, err
	}
	return built(configuration)
}

func built(configuration file) (Policy, error) {
	denied, err := sideFrom(configuration.Denied, asWritten)
	if err != nil {
		return Policy{}, err
	}
	allowed, err := sideFrom(configuration.Allowed, vouching(configuration.Trusted))
	return Policy{denied: denied, allowed: allowed}, err
}

func sideFrom(entries []entry, read func(rules.Rule) rules.Rule) (side, error) {
	commandRules, err := group(entries, read)
	if err != nil {
		return side{}, err
	}
	pathRules, err := globs(entries)
	return side{commandRules: commandRules, pathRules: pathRules}, err
}

func globs(entries []entry) (rules.Globs, error) {
	var found rules.Globs
	for _, written := range entries {
		list, err := entryGlobs(written)
		if err != nil {
			return nil, err
		}
		found = append(found, list...)
	}
	return found, nil
}

func entryGlobs(written entry) (rules.Globs, error) {
	switch {
	case len(written.Paths) == 0:
		return nil, nil
	case len(written.Commands) > 0 || len(written.Patterns) > 0:
		return nil, fmt.Errorf("an entry names `paths` or commands, never both: %+v", written)
	case written.Only != nil || len(written.Addendum) > 0:
		return nil, fmt.Errorf("`only` and `addendum` scope commands, and an entry with `paths` has none to scope: %+v", written)
	default:
		return globsIn(written)
	}
}

func globsIn(written entry) (rules.Globs, error) {
	var found rules.Globs
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

func group(entries []entry, read func(rules.Rule) rules.Rule) (rules.Group, error) {
	var found []rules.Rule
	for _, written := range entries {
		list, err := entryRules(written, read)
		if err != nil {
			return rules.Group{}, err
		}
		found = append(found, list...)
	}
	return rules.Group(found), nil
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
	if only == nil {
		return found, nil
	} else {
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

func allRules(written entry, read func(rules.Rule) rules.Rule) ([]rules.Rule, error) {
	if len(written.Commands) == 0 && len(written.Patterns) == 0 && len(written.Paths) == 0 {
		return nil, fmt.Errorf("an entry is a command, or a mapping with `commands`, `patterns` or `paths`: %+v", written)
	} else {
		return rulesIn(written, read)
	}
}

func rulesIn(written entry, read func(rules.Rule) rules.Rule) ([]rules.Rule, error) {
	var found []rules.Rule
	for _, text := range written.Commands {
		found = append(found, read(commandRule(text, written.Reason, written.Name)))
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

func commandRule(text, reason, name string) rules.Rule {
	if invocation, named := gitInvocation(text); named {
		return rules.NewGitSubcommand(invocation[0], invocation[1:], reason, name)
	} else {
		return rules.NewWords(text, reason, name)
	}
}

func gitInvocation(text string) (spoken []string, named bool) {
	if words := strings.Fields(text); len(words) < 2 || words[0] != "git" {
		return nil, false
	} else {
		return words[1:], true
	}
}
