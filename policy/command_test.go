// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"strings"
	"testing"

	"github.com/thehale/tooluse-screener/internal/rules"
)

var madeUp = Policy{
	denied: side{commandRules: rules.Group{
		pattern("sudo", "", ""),
		pattern("shutdown", "Ask a human first.", ""),
	}},
	allowed: side{commandRules: rules.Group{
		rules.NewOpening(rules.NewWords("ls", "", "")),
		rules.NewOpening(rules.NewWords("cat", "", "")),
	}},
}

func TestTheThreeAnswers(t *testing.T) {
	t.Run("a command matching an allowed rule", func(t *testing.T) {
		decides(t, Allow, "ls -la")
	})

	t.Run("a command matching a denied rule", func(t *testing.T) {
		decides(t, Deny, "sudo whoami")
	})

	t.Run("a command matching neither", func(t *testing.T) {
		decides(t, Ask, "nmap localhost")
	})

	t.Run("no command at all", func(t *testing.T) {
		decides(t, Ask, "   ")
	})
}

func TestAcrossACommandLine(t *testing.T) {
	t.Run("every command allowed allows the line", func(t *testing.T) {
		decides(t, Allow, "ls && cat foo")
	})

	t.Run("one unvetted command leaves the line to be asked about", func(t *testing.T) {
		decides(t, Ask, "ls && nmap localhost")
	})

	t.Run("one denied command denies the line", func(t *testing.T) {
		decides(t, Deny, "ls && sudo whoami")
	})

	t.Run("a denied command inside a substitution denies the line", func(t *testing.T) {
		decides(t, Deny, `ls "$(sudo whoami)"`)
	})

	t.Run("a denial outranks a permission", func(t *testing.T) {
		both := Policy{
			denied:  side{commandRules: rules.Group{pattern("ls", "", "")}},
			allowed: side{commandRules: rules.Group{rules.NewOpening(rules.NewWords("ls", "", ""))}},
		}
		if decision := both.CheckCommand("ls").Decision; decision != Deny {
			t.Errorf("ls -> %s, wanted deny", decision)
		}
	})
}

func TestReasons(t *testing.T) {
	t.Run("a refusal names the rule", func(t *testing.T) {
		ends(t, "sudo whoami", "sudo")
	})

	t.Run("a refusal carries the rule's reason", func(t *testing.T) {
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
		"denied:\n  - name: Push\n    commands: git push\n"
	landing := denial + "allowed:\n  - name: Push where CI can run it\n" +
		"    patterns: ['^git -C \\S+ push origin \\S+$']\n" +
		"    only:\n      dirs: [" + root + "]\n      branches: [{not: main}]\n"

	t.Run("a permission accounting for more of the command outranks the denial", func(t *testing.T) {
		answers(t, Allow, policyFrom(t, landing), "git -C "+root+" push origin topic")
	})

	t.Run("and the denial answers everywhere that permission does not", func(t *testing.T) {
		vouched := policyFrom(t, landing)
		answers(t, Deny, vouched, "git -C "+root+" push origin main")
		answers(t, Deny, vouched, "git -C "+root+" push")
		answers(t, Deny, vouched, "git -C /elsewhere push origin topic")
	})

	t.Run("a permission accounting for less does not outrank it", func(t *testing.T) {
		answers(t, Deny, policyFrom(t, denial+"allowed: [git]\n"), "git push origin topic")
	})

	t.Run("a tie goes to the denial", func(t *testing.T) {
		answers(t, Deny, policyFrom(t, denial+"allowed: [git push]\n"), "git push")
	})

	t.Run("and the same measure runs the other way", func(t *testing.T) {
		removing := policyFrom(t, "allowed: [rm]\ndenied:\n  - patterns: ['^rm -(rf|fr) /$']\n")
		answers(t, Deny, removing, "rm -rf /")
		answers(t, Allow, removing, "rm ./notes.txt")
	})
}

func TestAnAddendum(t *testing.T) {
	root := t.TempDir()
	chosen := policyFrom(t, "denied:\n  - commands: git push\n    reason: Ask first.\n"+
		"    addendum:\n      - only: {dirs: ["+root+"]}\n        reason: Push your own branch by name.\n")

	t.Run("ends the reason where the command runs inside its dirs", func(t *testing.T) {
		refusesWith(t, chosen, "git -C "+root+" push", "Ask first. Push your own branch by name.")
	})

	t.Run("is left out elsewhere", func(t *testing.T) {
		refusesWith(t, chosen, "git -C /elsewhere push", "Ask first.")
	})
}

func TestAnEmptyPolicy(t *testing.T) {
	t.Run("decides nothing", func(t *testing.T) {
		if decision := (Policy{}).CheckCommand("anything at all").Decision; decision != Ask {
			t.Errorf("wanted ask, got %s", decision)
		}
	})
}

func pattern(expression, reason, name string) rules.Rule {
	rule, err := rules.NewPattern(expression, reason, name)
	if err != nil {
		panic(err)
	}
	return rule
}

func answers(t *testing.T, wanted Decision, chosen Policy, command string) {
	t.Helper()
	if verdict := chosen.CheckCommand(command); verdict.Decision != wanted {
		t.Errorf("%q -> %s, wanted %s", command, verdict, wanted)
	}
}

func decides(t *testing.T, wanted Decision, command string) {
	t.Helper()
	if verdict := madeUp.CheckCommand(command); verdict.Decision != wanted {
		t.Errorf("%q -> %s, wanted %s", command, verdict, wanted)
	}
}

func refusesWith(t *testing.T, chosen Policy, command, wanted string) {
	t.Helper()
	if reason := chosen.CheckCommand(command).Reason; !strings.HasSuffix(reason, ": git push. "+wanted) {
		t.Errorf("%q -> %q, wanted it to end with %q", command, reason, wanted)
	}
}

func ends(t *testing.T, command, wanted string) {
	t.Helper()
	if reason := madeUp.CheckCommand(command).Reason; !strings.HasSuffix(reason, wanted) {
		t.Errorf("%q -> %q, wanted it to end with %q", command, reason, wanted)
	}
}
