// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

// Package check knows nothing about which commands are refused. That is
// a [policy.Policy].
package check

import (
	"fmt"
	"strings"

	"github.com/thehale/tooluse-screener/internal/command"
	"github.com/thehale/tooluse-screener/internal/lists"
	"github.com/thehale/tooluse-screener/internal/policy"
	"github.com/thehale/tooluse-screener/internal/rules"
)

type Decision string

const (
	Allow Decision = "allow"
	Deny  Decision = "deny"
	Ask   Decision = "ask"
)

type Verdict struct {
	Decision Decision
	Reason   string
}

func (v Verdict) String() string {
	return fmt.Sprintf("%s: %s", v.Decision, v.Reason)
}

// Evaluate denies a line for one denied command, and allows it only
// where every command is allowed. Denials are read first.
func Evaluate(line string, p policy.Policy) Verdict {
	found := command.All(line)
	refused := refusals(p.Denied, found)
	permitted := permissions(p.Allowed, found)

	switch {
	case len(refused) > 0:
		return Verdict{Deny, refusal(refused[0])}
	case len(found) > 0 && lists.Every(permitted, anyOf):
		return Verdict{Allow, permission(permitted)}
	default:
		return undecided
	}
}

var undecided = Verdict{Ask, "Command is not in the shared allow list"}

func refusals(denied rules.Group, found []string) []rules.Rule {
	var matched []rules.Rule
	for _, one := range found {
		matched = append(matched, denied.Matching(one)...)
	}
	return matched
}

func permissions(allowed rules.Group, found []string) [][]rules.Rule {
	vouching := make([][]rules.Rule, 0, len(found))
	for _, one := range found {
		vouching = append(vouching, allowed.Matching(one))
	}
	return vouching
}

func anyOf(matched []rules.Rule) bool {
	return len(matched) > 0
}

func refusal(rule rules.Rule) string {
	refused := fmt.Sprintf("Command matches a denied rule: %s", rule)
	if rule.Note() == "" {
		return refused
	}
	return fmt.Sprintf("%s. %s", refused, rule.Note())
}

func permission(permitted [][]rules.Rule) string {
	return fmt.Sprintf("Every command is allowed: %s", strings.Join(named(permitted), ", "))
}

func named(permitted [][]rules.Rule) []string {
	var spoken []string
	seen := map[string]bool{}
	for _, matched := range permitted {
		for _, rule := range matched {
			if name := rule.String(); !seen[name] {
				seen[name] = true
				spoken = append(spoken, name)
			}
		}
	}
	return spoken
}
