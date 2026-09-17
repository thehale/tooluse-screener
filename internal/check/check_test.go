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
		rules.NewOpening(rules.NewWords("ls", "", "")),
		rules.NewOpening(rules.NewWords("cat", "", "")),
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
			Allowed: rules.NewGroup("allowed", rules.NewOpening(rules.NewWords("ls", "", ""))),
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

func TestWhichRuleOutranksWhich(t *testing.T) {
	root := t.TempDir()
	denial := "trusted_git_directories: [" + root + "]\n" +
		"denied:\n  - description: Pushing\n    commands: git push\n"
	landing := denial + "allowed:\n  - description: Pushing where CI can run it\n" +
		"    patterns: ['^git -C \\S+ push origin \\S+$']\n" +
		"    only:\n      dirs: [" + root + "]\n      branches: [{not: main}]\n"

	t.Run("a permission accounting for more of the command outranks the denial", func(t *testing.T) {
		answers(t, check.Allow, parsed(t, landing), "git -C "+root+" push origin topic")
	})

	t.Run("and the denial answers everywhere that permission does not", func(t *testing.T) {
		vouched := parsed(t, landing)
		answers(t, check.Deny, vouched, "git -C "+root+" push origin main")
		answers(t, check.Deny, vouched, "git -C "+root+" push")
		answers(t, check.Deny, vouched, "git -C /elsewhere push origin topic")
	})

	t.Run("a permission accounting for less does not outrank it", func(t *testing.T) {
		answers(t, check.Deny, parsed(t, denial+"allowed: [git]\n"), "git push origin topic")
	})

	t.Run("a tie goes to the denial", func(t *testing.T) {
		answers(t, check.Deny, parsed(t, denial+"allowed: [git push]\n"), "git push")
	})

	t.Run("and the same measure runs the other way", func(t *testing.T) {
		removing := parsed(t, "allowed: [rm]\ndenied:\n  - patterns: ['^rm -(rf|fr) /$']\n")
		answers(t, check.Deny, removing, "rm -rf /")
		answers(t, check.Allow, removing, "rm ./notes.txt")
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

func parsed(t *testing.T, configuration string) policy.Policy {
	t.Helper()
	built, err := policy.Parse([]byte(configuration))
	if err != nil {
		t.Fatal(err)
	}
	return built
}

func answers(t *testing.T, wanted check.Decision, chosen policy.Policy, command string) {
	t.Helper()
	if verdict := check.Evaluate(command, chosen); verdict.Decision != wanted {
		t.Errorf("%q -> %s, wanted %s", command, verdict, wanted)
	}
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
