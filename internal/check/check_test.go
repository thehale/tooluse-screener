// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package check_test

import (
	"strings"
	"testing"

	"github.com/thehale/tooluse-screener/internal/check"
	"github.com/thehale/tooluse-screener/internal/policy"
	"github.com/thehale/tooluse-screener/internal/rules"
)

var madeUp = policy.Policy{
	Denied: rules.NewGroup("denied",
		pattern("sudo", "", ""),
		pattern("shutdown", "Ask a human first.", ""),
	),
	Allowed: rules.NewGroup("allowed",
		rules.NewPrefix("ls", "", ""),
		rules.NewPrefix("cat", "", ""),
	),
}

func TestTheThreeAnswers(t *testing.T) {
	t.Run("a command matching an allowed rule", func(t *testing.T) {
		decides(t, check.Allow, "ls -la")
	})

	t.Run("a command matching a denied rule", func(t *testing.T) {
		decides(t, check.Deny, "sudo whoami")
	})

	t.Run("a command matching neither", func(t *testing.T) {
		decides(t, check.Ask, "nmap localhost")
	})

	t.Run("no command at all", func(t *testing.T) {
		decides(t, check.Ask, "   ")
	})
}

func TestAcrossACommandLine(t *testing.T) {
	t.Run("every command allowed allows the line", func(t *testing.T) {
		decides(t, check.Allow, "ls && cat foo")
	})

	t.Run("one unvetted command leaves the line to be asked about", func(t *testing.T) {
		decides(t, check.Ask, "ls && nmap localhost")
	})

	t.Run("one denied command denies the line", func(t *testing.T) {
		decides(t, check.Deny, "ls && sudo whoami")
	})

	t.Run("a denied command inside a substitution denies the line", func(t *testing.T) {
		decides(t, check.Deny, `ls "$(sudo whoami)"`)
	})

	t.Run("a denial outranks a permission", func(t *testing.T) {
		both := policy.Policy{
			Denied:  rules.NewGroup("denied", pattern("ls", "", "")),
			Allowed: rules.NewGroup("allowed", rules.NewPrefix("ls", "", "")),
		}
		if decision := check.Evaluate("ls", both).Decision; decision != check.Deny {
			t.Errorf("ls -> %s, wanted deny", decision)
		}
	})
}

func TestReasons(t *testing.T) {
	t.Run("a refusal names the rule", func(t *testing.T) {
		ends(t, "sudo whoami", "sudo")
	})

	t.Run("a refusal carries the rule's note", func(t *testing.T) {
		ends(t, "shutdown now", "Ask a human first.")
	})

	t.Run("a permission names every rule that vouched", func(t *testing.T) {
		ends(t, "ls && cat foo", "ls, cat")
	})

	t.Run("a permission names a rule once", func(t *testing.T) {
		ends(t, "ls && ls -la", "ls")
	})
}

func TestAnEmptyPolicy(t *testing.T) {
	t.Run("decides nothing", func(t *testing.T) {
		if decision := check.Evaluate("anything at all", policy.Policy{}).Decision; decision != check.Ask {
			t.Errorf("wanted ask, got %s", decision)
		}
	})
}

func pattern(expression, note, description string) rules.Rule {
	rule, err := rules.NewPattern(expression, note, description)
	if err != nil {
		panic(err)
	}
	return rule
}

func decides(t *testing.T, wanted check.Decision, command string) {
	t.Helper()
	if verdict := check.Evaluate(command, madeUp); verdict.Decision != wanted {
		t.Errorf("%q -> %s, wanted %s", command, verdict, wanted)
	}
}

func ends(t *testing.T, command, wanted string) {
	t.Helper()
	if reason := check.Evaluate(command, madeUp).Reason; !strings.HasSuffix(reason, wanted) {
		t.Errorf("%q -> %q, wanted it to end with %q", command, reason, wanted)
	}
}
