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
	commands := command.All(line)
	reasons := p.refusals(commands)
	vouchersByCommand := permissions(p.allowed.commandRules, commands)

	switch {
	case len(reasons) > 0:
		return Verdict{Deny, reasons[0]}
	case len(commands) > 0 && lists.Every(vouchersByCommand, isVouchedFor):
		return Verdict{Allow, permission(vouchersByCommand)}
	default:
		return unlisted("Command")
	}
}

func (p Policy) refusals(commands []command.Command) []string {
	var reasons []string
	for _, one := range commands {
		for _, rule := range standingDenials(p.denied.commandRules.RulesMatching(one), p.allowed.commandRules, one) {
			reasons = append(reasons, refusal("Command", rule, rule.Reason(one)))
		}
	}
	return reasons
}

func standingDenials(denials rules.Group, allowRules rules.Group, line command.Command) rules.Group {
	var standing rules.Group
	for _, rule := range denials {
		if !line.IsAccountedForBy(allowRules) || line.IsAccountedForBy(rule) {
			standing = append(standing, rule)
		}
	}
	return standing
}

func permissions(allowRules rules.Group, commands []command.Command) []rules.Group {
	vouchersByCommand := make([]rules.Group, 0, len(commands))
	for _, one := range commands {
		vouchersByCommand = append(vouchersByCommand, allowRules.RulesMatching(one))
	}
	return vouchersByCommand
}

func isVouchedFor(vouchers rules.Group) bool {
	return len(vouchers) > 0
}

func permission(vouchersByCommand []rules.Group) string {
	return fmt.Sprintf("Every command is allowed: %s", strings.Join(names(vouchersByCommand), ", "))
}

func names(vouchersByCommand []rules.Group) []string {
	var ruleNames []string
	nameSet := map[string]bool{}
	for _, vouchers := range vouchersByCommand {
		for _, rule := range vouchers {
			if name := rule.String(); !nameSet[name] {
				nameSet[name] = true
				ruleNames = append(ruleNames, name)
			}
		}
	}
	return ruleNames
}
