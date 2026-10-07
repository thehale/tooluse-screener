// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command_test

import (
	"testing"

	"github.com/thehale/tooluse-screener/internal/command"
)

type spanning int

func (s spanning) Span(command.Command) int {
	return int(s)
}

func TestAccounting(t *testing.T) {
	t.Run("a rule spanning every spoken word accounts for the command", func(t *testing.T) {
		accountsFor(t, true, "git status", spanning(len("git status")))
		accountsFor(t, true, "GIT_PAGER=cat git status", spanning(len("git status")))
	})

	t.Run("a rule spanning less does not", func(t *testing.T) {
		accountsFor(t, false, "git status --short", spanning(len("git status")))
	})

	t.Run("nothing spoken is accounted for by nothing", func(t *testing.T) {
		accountsFor(t, false, "GIT_PAGER=cat", spanning(0))
	})
}

func accountsFor(t *testing.T, wanted bool, text string, rule command.Spanner) {
	t.Helper()
	if got := (command.Command{Text: text}).IsAccountedForBy(rule); got != wanted {
		t.Errorf("%q accounted for = %v, wanted %v", text, got, wanted)
	}
}
