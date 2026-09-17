// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package rules_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/thehale/tooluse-screener/internal/rules"
)

func TestSubstring(t *testing.T) {
	deploying := rules.NewSubstring("deploy now", "", "")

	t.Run("matches a command containing it", func(t *testing.T) {
		matches(t, true, deploying, "deploy now")
	})

	t.Run("matches it behind anything", func(t *testing.T) {
		matches(t, true, deploying, "RELEASE_CHANNEL=live deploy now")
	})

	t.Run("does not match otherwise", func(t *testing.T) {
		matches(t, false, deploying, "echo hi")
	})

	t.Run("its text is literal", func(t *testing.T) {
		matches(t, true, rules.NewSubstring("a.c", "", ""), "a.c")
		matches(t, false, rules.NewSubstring("a.c", "", ""), "abc")
	})

	t.Run("an empty rule matches nothing", func(t *testing.T) {
		matches(t, false, rules.NewSubstring("", "", ""), "anything")
	})
}

func TestPattern(t *testing.T) {
	never := pattern(t, `\bnever\s+ever\b`, "", "")

	t.Run("matches a command it matches", func(t *testing.T) {
		matches(t, true, never, "do never   ever run this")
	})

	t.Run("does not match otherwise", func(t *testing.T) {
		matches(t, false, never, "never mind")
	})

	t.Run("it names itself with its expression", func(t *testing.T) {
		named(t, pattern(t, `\bnever\b`, "", ""), `\bnever\b`)
	})

	t.Run("an expression that will not compile is refused", func(t *testing.T) {
		if _, err := rules.NewPattern("([", "", ""); err == nil {
			t.Error("wanted an error")
		}
		if _, err := rules.NewLeading("([", "", ""); err == nil {
			t.Error("wanted an error")
		}
	})
}

func TestPrefix(t *testing.T) {
	listing := rules.NewPrefix("ls", "", "")

	t.Run("matches the bare command", func(t *testing.T) {
		matches(t, true, listing, "ls")
	})

	t.Run("matches the command with arguments", func(t *testing.T) {
		matches(t, true, listing, "ls -la")
	})

	t.Run("does not match a longer word", func(t *testing.T) {
		matches(t, false, listing, "lsblk")
	})

	t.Run("does not match further along", func(t *testing.T) {
		matches(t, false, listing, "echo ls")
	})

	t.Run("its text is literal", func(t *testing.T) {
		matches(t, false, rules.NewPrefix("ls|cat", "", ""), "cat foo")
	})

	t.Run("an empty rule matches nothing", func(t *testing.T) {
		matches(t, false, rules.NewPrefix("", "", ""), "anything")
	})
}

func TestLeading(t *testing.T) {
	t.Run("matches from the start", func(t *testing.T) {
		matches(t, true, leading(t, "ls|cat", "", ""), "cat foo")
	})

	t.Run("does not match further along", func(t *testing.T) {
		matches(t, false, leading(t, "ls|cat", "", ""), "echo cat")
	})

	t.Run("ends at a word boundary", func(t *testing.T) {
		matches(t, false, leading(t, "ls", "", ""), "lsblk")
	})
}

func TestGitSubcommand(t *testing.T) {
	pushing := rules.NewGitSubcommand("push", nil, "", "")

	t.Run("matches the plain spelling", func(t *testing.T) {
		matches(t, true, pushing, "git push origin main")
	})

	t.Run("matches the -C spelling", func(t *testing.T) {
		matches(t, true, pushing, "git -C /anywhere push")
	})

	t.Run("does not match another subcommand", func(t *testing.T) {
		matches(t, false, pushing, "git status")
	})

	t.Run("does not match a command that only mentions it", func(t *testing.T) {
		matches(t, false, pushing, "grep -rn 'git push' docs")
	})

	t.Run("does not match a bare git", func(t *testing.T) {
		matches(t, false, pushing, "git")
	})

	t.Run("names itself with the word git", func(t *testing.T) {
		named(t, pushing, "git push")
	})
}

func TestGitSubcommandArguments(t *testing.T) {
	unsetting := rules.NewGitSubcommand("config", []string{"--unset"}, "", "")

	t.Run("matches the subcommand carrying them", func(t *testing.T) {
		matches(t, true, unsetting, "git config --unset user.name")
	})

	t.Run("matches with another flag in between", func(t *testing.T) {
		matches(t, true, unsetting, "git config --global --unset user.name")
	})

	t.Run("does not match without them", func(t *testing.T) {
		matches(t, false, unsetting, "git config --get user.name")
	})

	t.Run("names itself with its arguments", func(t *testing.T) {
		named(t, unsetting, "git config --unset")
	})
}

func TestGitSubcommandDirectories(t *testing.T) {
	root := t.TempDir()
	within := func(roots ...string) rules.Rule {
		return rules.NewGitSubcommandWithin("status", nil, roots, "", "")
	}

	t.Run("without within it matches wherever it points", func(t *testing.T) {
		matches(t, true, rules.NewGitSubcommand("push", nil, "", ""), "git -C /anywhere/at/all push")
	})

	t.Run("with within it matches a trusted directory", func(t *testing.T) {
		matches(t, true, within(root), "git -C "+root+" status")
	})

	t.Run("with within it matches a subdirectory of one", func(t *testing.T) {
		matches(t, true, within(root), "git -C "+filepath.Join(root, "nested")+" status")
	})

	t.Run("with within it does not match a sibling", func(t *testing.T) {
		matches(t, false, within(filepath.Join(root, "foo")), "git -C "+filepath.Join(root, "foobar")+" status")
	})

	t.Run("one untrusted directory spoils the command", func(t *testing.T) {
		trusted := filepath.Join(root, "trusted")
		matches(t, false, within(trusted), "git -C "+trusted+" -C "+filepath.Join(root, "outside")+" status")
	})

	t.Run("a directory written with an equals sign", func(t *testing.T) {
		matches(t, true, within(root), "git -C="+root+" status")
	})

	t.Run("a relative path is resolved", func(t *testing.T) {
		matches(t, true, within(root), "git -C "+filepath.Join(root, "nested", "..", "nested")+" status")
	})

	t.Run("a tilde root is expanded", func(t *testing.T) {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skip("no home directory")
		}
		matches(t, true, within("~"), "git -C "+home+" status")
	})

	t.Run("with within it still matches a command with no directory", func(t *testing.T) {
		matches(t, true, within(root), "git status")
	})

	t.Run("bound to no roots at all it matches nothing pointed anywhere", func(t *testing.T) {
		matches(t, true, within(), "git status")
		matches(t, false, within(), "git -C "+root+" status")
	})
}

func TestGroup(t *testing.T) {
	listing, reading := rules.NewPrefix("ls", "", ""), rules.NewPrefix("cat", "", "")

	t.Run("matches when one of its rules does", func(t *testing.T) {
		matches(t, true, rules.NewGroup("", listing, reading), "cat foo")
	})

	t.Run("does not match when none of them does", func(t *testing.T) {
		matches(t, false, rules.NewGroup("", listing), "nmap localhost")
	})

	t.Run("an empty group matches nothing", func(t *testing.T) {
		matches(t, false, rules.NewGroup(""), "anything")
	})

	t.Run("it lists the rules that matched rather than itself", func(t *testing.T) {
		matched(t, rules.NewGroup("", listing, reading), "cat foo", reading)
	})

	t.Run("it lists every rule that matched", func(t *testing.T) {
		narrow := rules.NewPrefix("ls -la", "", "")
		matched(t, rules.NewGroup("", listing, narrow), "ls -la", listing, narrow)
	})

	t.Run("it lists nothing when none matched", func(t *testing.T) {
		matched(t, rules.NewGroup("", listing), "nmap localhost")
	})

	t.Run("a nested group lists the innermost rules", func(t *testing.T) {
		matched(t, rules.NewGroup("", rules.NewGroup("", listing)), "ls -la", listing)
	})

	t.Run("it names itself", func(t *testing.T) {
		named(t, rules.NewGroup("reading"), "reading")
	})
}

func TestWhatARuleSaysAboutItself(t *testing.T) {
	t.Run("a description is what a rule calls itself", func(t *testing.T) {
		named(t, pattern(t, `\bx\b`, "", "Something unreadable"), "Something unreadable")
	})

	t.Run("without one it calls itself what it is written as", func(t *testing.T) {
		named(t, pattern(t, `\bx\b`, "", ""), `\bx\b`)
	})

	t.Run("a note is advice, kept apart from the name", func(t *testing.T) {
		rule := rules.NewPrefix("ls", "Ask a human first.", "")
		if rule.Note() != "Ask a human first." {
			t.Errorf("Note() = %q", rule.Note())
		}
		named(t, rule, "ls")
	})
}

func pattern(t *testing.T, expression, note, description string) rules.Rule {
	t.Helper()
	rule, err := rules.NewPattern(expression, note, description)
	if err != nil {
		t.Fatal(err)
	}
	return rule
}

func leading(t *testing.T, expression, note, description string) rules.Rule {
	t.Helper()
	rule, err := rules.NewLeading(expression, note, description)
	if err != nil {
		t.Fatal(err)
	}
	return rule
}

func matches(t *testing.T, wanted bool, rule rules.Rule, command string) {
	t.Helper()
	if got := rule.Matches(command); got != wanted {
		t.Errorf("%s matching %q = %v, wanted %v", rule, command, got, wanted)
	}
}

func named(t *testing.T, rule rules.Rule, wanted string) {
	t.Helper()
	if got := rule.String(); got != wanted {
		t.Errorf("String() = %q, wanted %q", got, wanted)
	}
}

func matched(t *testing.T, group rules.Group, command string, wanted ...rules.Rule) {
	t.Helper()
	if got := group.Matching(command); !slices.Equal(got, wanted) {
		t.Errorf("Matching(%q) = %v, wanted %v", command, got, wanted)
	}
}
