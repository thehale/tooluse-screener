// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"fmt"
	"strings"

	"github.com/thehale/tooluse-screener/internal/command"
	"github.com/thehale/tooluse-screener/internal/lists"
	"github.com/thehale/tooluse-screener/internal/rules"
)

func (p Policy) CheckCommand(line string) Verdict {
	found := command.All(line)
	refused := p.refusals(found)
	permitted := permissions(p.allowed.commandRules, found)

	switch {
	case len(refused) > 0:
		return Verdict{Deny, refused[0]}
	case len(found) > 0 && lists.Every(permitted, isVouchedFor):
		return Verdict{Allow, permission(permitted)}
	default:
		return unlistedCommand
	}
}

var unlistedCommand = Verdict{Ask, "Command is not in the shared allow list"}

func (p Policy) refusals(found []command.Command) []string {
	var reasons []string
	for _, one := range found {
		for _, rule := range standingDenials(p.denied.commandRules.Matching(one), p.allowed.commandRules, one) {
			reasons = append(reasons, refusal("Command", rule, rule.Reason(one)))
		}
	}
	return reasons
}

func standingDenials(denials []rules.Rule, allowed rules.Group, line command.Command) []rules.Rule {
	var matched []rules.Rule
	for _, rule := range denials {
		if !isAccountedForBy(line, allowed) || isAccountedForBy(line, rule) {
			matched = append(matched, rule)
		}
	}
	return matched
}

type spanning interface {
	Span(command command.Command) int
}

func isAccountedForBy(line command.Command, rule spanning) bool {
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

func isVouchedFor(matched []rules.Rule) bool {
	return len(matched) > 0
}

func refusal(subject string, rule fmt.Stringer, reason string) string {
	refused := fmt.Sprintf("%s matches a denied rule: %s", subject, rule)
	if reason == "" {
		return refused
	} else {
		return fmt.Sprintf("%s. %s", refused, reason)
	}
}

func permission(permitted [][]rules.Rule) string {
	return fmt.Sprintf("Every command is allowed: %s", strings.Join(names(permitted), ", "))
}

func names(permitted [][]rules.Rule) []string {
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
