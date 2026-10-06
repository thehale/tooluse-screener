// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

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

func Evaluate(line string, p policy.Policy) Verdict {
	found := command.All(line)
	refused := refusals(p, found)
	permitted := permissions(p.Allowed, found)

	switch {
	case len(refused) > 0:
		return Verdict{Deny, refused[0]}
	case len(found) > 0 && lists.Every(permitted, anyOf):
		return Verdict{Allow, permission(permitted)}
	default:
		return undecided
	}
}

var undecided = Verdict{Ask, "Command is not in the shared allow list"}

func refusals(p policy.Policy, found []command.Command) []string {
	var reasons []string
	for _, one := range found {
		for _, rule := range standing(p.Denied.Matching(one), p.Allowed, one) {
			reasons = append(reasons, refusal("Command", rule, rule.Note(one)))
		}
	}
	return reasons
}

func standing(denials []rules.Rule, allowed rules.Group, line command.Command) []rules.Rule {
	var matched []rules.Rule
	for _, rule := range denials {
		if !accountsForAll(allowed, line) || accountsForAll(rule, line) {
			matched = append(matched, rule)
		}
	}
	return matched
}

func accountsForAll(rule rules.Rule, line command.Command) bool {
	spoken := command.Spoken(line.Text)
	return spoken != "" && rule.Span(line) == len(spoken)
}

func permissions(allowed rules.Group, found []command.Command) [][]rules.Rule {
	vouching := make([][]rules.Rule, 0, len(found))
	for _, one := range found {
		vouching = append(vouching, allowed.Matching(one))
	}
	return vouching
}

func anyOf(matched []rules.Rule) bool {
	return len(matched) > 0
}

func refusal(subject string, rule fmt.Stringer, note string) string {
	refused := fmt.Sprintf("%s matches a denied rule: %s", subject, rule)
	if note == "" {
		return refused
	}
	return fmt.Sprintf("%s. %s", refused, note)
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
