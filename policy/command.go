// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"fmt"
	"slices"
	"strings"

	"github.com/thehale/tooluse-screener/internal/command"
	"github.com/thehale/tooluse-screener/internal/lists"
	"github.com/thehale/tooluse-screener/internal/rules"
)

func (p Policy) CheckCommand(line string) Verdict {
	commands := command.All(line)
	reasons := p.denyReasons(slices.Concat(commands, command.Blind(line)))
	matchingByCommand := allowRulesByCommand(p.allowed.commandRules, commands)

	switch {
	case len(reasons) > 0:
		return Verdict{Deny, reasons[0]}
	case len(commands) > 0 && lists.Every(matchingByCommand, isAllowed):
		return Verdict{Allow, allowReason(matchingByCommand)}
	default:
		return unlistedVerdict("Command")
	}
}

func (p Policy) denyReasons(commands []command.Command) []string {
	var reasons []string
	for _, one := range commands {
		for _, rule := range standingDenyRules(p.denied.commandRules.RulesMatching(one), p.allowed.commandRules, one) {
			reasons = append(reasons, denyReason("Command", rule, rule.Reason(one)))
		}
	}
	return reasons
}

func standingDenyRules(denyRules rules.Group, allowRules rules.Group, line command.Command) rules.Group {
	var standingRules rules.Group
	for _, rule := range denyRules {
		if !line.IsAccountedForBy(allowRules) || line.IsAccountedForBy(rule) {
			standingRules = append(standingRules, rule)
		}
	}
	return standingRules
}

func allowRulesByCommand(allowRules rules.Group, commands []command.Command) []rules.Group {
	matchingByCommand := make([]rules.Group, 0, len(commands))
	for _, one := range commands {
		matchingByCommand = append(matchingByCommand, allowRules.RulesMatching(one))
	}
	return matchingByCommand
}

func isAllowed(matchingRules rules.Group) bool {
	return len(matchingRules) > 0
}

func allowReason(matchingByCommand []rules.Group) string {
	return fmt.Sprintf("Every command is allowed: %s", strings.Join(names(matchingByCommand), ", "))
}

func names(matchingByCommand []rules.Group) []string {
	var ruleNames []string
	nameSet := map[string]bool{}
	for _, matchingRules := range matchingByCommand {
		for _, rule := range matchingRules {
			if name := rule.String(); !nameSet[name] {
				nameSet[name] = true
				ruleNames = append(ruleNames, name)
			}
		}
	}
	return ruleNames
}
