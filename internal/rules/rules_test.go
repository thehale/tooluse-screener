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

func TestWords(t *testing.T) {
	deploying := rules.NewWords("deploy now", "", "")

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
		matches(t, true, rules.NewWords("a.c", "", ""), "a.c")
		matches(t, false, rules.NewWords("a.c", "", ""), "abc")
	})

	t.Run("an empty rule matches nothing", func(t *testing.T) {
		matches(t, false, rules.NewWords("", "", ""), "anything")
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
	})
}

func TestOpening(t *testing.T) {
	listing := rules.NewOpening(rules.NewWords("ls", "", ""))

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
		matches(t, false, rules.NewOpening(rules.NewWords("ls|cat", "", "")), "cat foo")
	})

	t.Run("an empty rule matches nothing", func(t *testing.T) {
		matches(t, false, rules.NewOpening(rules.NewWords("", "", "")), "anything")
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
		matches(t, false, leading(t, "ls|cat", "", ""), "catalog foo")
	})

	t.Run("an expression anchored to the end means the end of the command", func(t *testing.T) {
		exactly := leading(t, `echo \S+$`, "", "")
		matches(t, true, exactly, "echo hi")
		matches(t, false, exactly, "echo hi there")
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

func TestReaching(t *testing.T) {
	root := t.TempDir()
	within := func(roots ...string) rules.Rule {
		return rules.NewReaching(rules.NewGitSubcommand("status", nil, "", ""), roots)
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

	t.Run("a repository named instead of entered is judged the same way", func(t *testing.T) {
		matches(t, true, within(root), "git --git-dir="+filepath.Join(root, ".git")+" status")
		matches(t, false, within(root), "git --git-dir=/elsewhere/.git status")
		matches(t, false, within(root), "git --work-tree /elsewhere status")
	})

	t.Run("with within it still matches a command with no directory", func(t *testing.T) {
		matches(t, true, within(root), "git status")
	})

	t.Run("bound to no roots at all it matches nothing pointed anywhere", func(t *testing.T) {
		matches(t, true, within(), "git status")
		matches(t, false, within(), "git -C "+root+" status")
	})
}

func TestRestricted(t *testing.T) {
	root := t.TempDir()
	pushing := rules.NewGitSubcommand("push", nil, "Open a pull request.", "")
	onto := func(heldBack ...string) rules.Rule {
		return rules.NewRestricted(pushing, []string{root}, rules.Branches{NotOnto: heldBack})
	}

	t.Run("matches a branch the command names", func(t *testing.T) {
		matches(t, true, onto("main"), "git -C "+root+" push origin topic")
	})

	t.Run("does not match one held back, however it is written", func(t *testing.T) {
		matches(t, false, onto("main"), "git -C "+root+" push origin main")
		matches(t, false, onto("main"), "git -C "+root+" push origin HEAD:main")
		matches(t, false, onto("main"), "git -C "+root+" push origin MAIN")
	})

	t.Run("does not match a spelling that hides where it lands", func(t *testing.T) {
		matches(t, false, onto("main"), "git -C "+root+" push")
		matches(t, false, onto("main"), "git -C "+root+" push --force origin topic")
	})

	t.Run("does not match outside the dirs it is given", func(t *testing.T) {
		matches(t, false, onto("main"), "git -C /elsewhere push origin topic")
		matches(t, false, rules.NewRestricted(pushing, []string{"/nowhere"}, rules.Branches{}), "git -C "+root+" push origin topic")
	})

	t.Run("a command naming no directory is judged where it runs", func(t *testing.T) {
		here, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		matches(t, true, rules.NewRestricted(pushing, []string{here}, rules.Branches{}), "git push origin topic")
		matches(t, false, rules.NewRestricted(pushing, []string{root}, rules.Branches{}), "git push origin topic")
	})

	t.Run("an unrestricted rule is left as it was", func(t *testing.T) {
		matches(t, true, rules.NewRestricted(pushing, nil, rules.Branches{}), "git -C /anywhere push")
	})

	t.Run("speaks for the rule it restricts", func(t *testing.T) {
		named(t, onto("main"), "git push")
		if onto("main").Note() != "Open a pull request." {
			t.Errorf("Note() = %q", onto("main").Note())
		}
	})
}

func TestHowMuchOfACommandARuleAccountsFor(t *testing.T) {
	t.Run("a rule that does not match accounts for nothing", func(t *testing.T) {
		spans(t, 0, rules.NewOpening(rules.NewWords("ls", "", "")), "cat foo")
		spans(t, 0, pattern(t, `\bnever\b`, "", ""), "always")
	})

	t.Run("a text rule accounts for its own text", func(t *testing.T) {
		spans(t, 2, rules.NewOpening(rules.NewWords("ls", "", "")), "ls -la")
		spans(t, 13, rules.NewWords("chezmoi apply", "", ""), "sudo chezmoi apply --force")
	})

	t.Run("an expression accounts for what it matched", func(t *testing.T) {
		spans(t, 7, pattern(t, `ls|cat foo`, "", ""), "cat foo bar")
		spans(t, 3, leading(t, "ls|cat", "", ""), "cat foo")
	})

	t.Run("a git subcommand accounts for the command through the words it requires", func(t *testing.T) {
		spans(t, 8, rules.NewGitSubcommand("push", nil, "", ""), "git push origin topic")
		spans(t, 22, rules.NewGitSubcommand("push", nil, "", ""), "git -C /elsewhere push origin topic")
		spans(t, 18, rules.NewGitSubcommand("config", []string{"--unset"}, "", ""), "git config --unset user.name")
	})

	t.Run("a restricted rule accounts for what the rule it restricts did", func(t *testing.T) {
		here, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		naming := leading(t, `git push origin \S+$`, "", "")
		landing := rules.NewRestricted(naming, []string{here}, rules.Branches{NotOnto: []string{"main"}})
		spans(t, len("git push origin topic"), landing, "git push origin topic")
		spans(t, 0, landing, "git push origin main")
	})

	t.Run("a group accounts for the widest of its rules", func(t *testing.T) {
		both := rules.NewGroup("", rules.NewOpening(rules.NewWords("git", "", "")), rules.NewOpening(rules.NewWords("git push", "", "")))
		spans(t, 8, both, "git push origin topic")
	})
}

func TestGroup(t *testing.T) {
	listing, reading := rules.NewOpening(rules.NewWords("ls", "", "")), rules.NewOpening(rules.NewWords("cat", "", ""))

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
		narrow := rules.NewOpening(rules.NewWords("ls -la", "", ""))
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
		rule := rules.NewOpening(rules.NewWords("ls", "Ask a human first.", ""))
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
	return rules.NewOpening(pattern(t, expression, note, description))
}

func matches(t *testing.T, wanted bool, rule rules.Rule, command string) {
	t.Helper()
	if got := rule.Matches(command); got != wanted {
		t.Errorf("%s matching %q = %v, wanted %v", rule, command, got, wanted)
	}
}

func spans(t *testing.T, wanted int, rule rules.Rule, command string) {
	t.Helper()
	if got := rule.Span(command); got != wanted {
		t.Errorf("%s spanning %q = %d, wanted %d", rule, command, got, wanted)
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
