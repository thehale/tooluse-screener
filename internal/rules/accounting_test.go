// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules_test

import (
	"testing"

	"github.com/thehale/tooluse-screener/internal/command"
	"github.com/thehale/tooluse-screener/internal/rules"
)

func TestAccounting(t *testing.T) {
	spanningGitStatus := rules.Rule{Expression: rules.AtStart{Expression: rules.Words("git status")}}

	t.Run("a rule spanning every spoken word accounts for the command", func(t *testing.T) {
		accountsFor(t, true, "git status", spanningGitStatus)
		accountsFor(t, true, "GIT_PAGER=cat git status", spanningGitStatus)
	})

	t.Run("a rule spanning less does not", func(t *testing.T) {
		accountsFor(t, false, "git status --short", spanningGitStatus)
	})

	t.Run("nothing spoken is accounted for by nothing", func(t *testing.T) {
		accountsFor(t, false, "GIT_PAGER=cat", rules.Rule{Expression: rules.AtStart{Expression: rules.Words("")}})
	})
}

func accountsFor(t *testing.T, wanted bool, text string, rule rules.Rule) {
	t.Helper()
	if got := rule.AccountsFor(command.Command{Text: text}); got != wanted {
		t.Errorf("%q accounted for = %v, wanted %v", text, got, wanted)
	}
}
