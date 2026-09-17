// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thehale/tooluse-screener/internal/policy"
	"github.com/thehale/tooluse-screener/internal/rules"
)

func TestWhereACommandIsLookedFor(t *testing.T) {
	t.Run("a denied command is looked for anywhere", func(t *testing.T) {
		denied := parsed(t, "denied:\n  - deploy now\n").Denied
		matches(t, true, denied, "deploy now")
		matches(t, true, denied, "RELEASE_CHANNEL=live deploy now")
	})

	t.Run("an allowed command only at the start", func(t *testing.T) {
		allowed := parsed(t, "allowed:\n  - ls\n").Allowed
		matches(t, true, allowed, "ls -la")
		matches(t, false, allowed, "lsblk")
		matches(t, false, allowed, "echo ls")
	})

	t.Run("a command is literal text", func(t *testing.T) {
		denied := parsed(t, "denied:\n  - a.c\n").Denied
		matches(t, true, denied, "a.c")
		matches(t, false, denied, "abc")
	})
}

func TestWhereAPatternIsLookedFor(t *testing.T) {
	t.Run("a denied pattern is looked for anywhere", func(t *testing.T) {
		denied := parsed(t, "denied:\n  - patterns: ['\\bnever\\b']\n").Denied
		matches(t, true, denied, "do never run this")
	})

	t.Run("an allowed pattern only at the start", func(t *testing.T) {
		allowed := parsed(t, "allowed:\n  - patterns: ['ls|cat']\n").Allowed
		matches(t, true, allowed, "cat foo")
		matches(t, false, allowed, "echo cat")
	})
}

func TestACommandStartingWithGit(t *testing.T) {
	t.Run("is read as a git command", func(t *testing.T) {
		denied := parsed(t, "denied:\n  - git push\n").Denied
		matches(t, true, denied, "git push origin main")
		matches(t, true, denied, "git -C /anywhere push")
		matches(t, false, denied, "git status")
	})

	t.Run("carries its arguments", func(t *testing.T) {
		denied := parsed(t, "denied:\n  - git config --unset\n").Denied
		matches(t, true, denied, "git config --unset user.name")
		matches(t, true, denied, "git -C /x config --global --unset user.name")
		matches(t, false, denied, "git config --get user.name")
	})

	t.Run("is not looked for as text", func(t *testing.T) {
		denied := parsed(t, "denied:\n  - git push\n").Denied
		matches(t, false, denied, "grep -rn 'git push' docs")
	})

	t.Run("a bare git is left as text", func(t *testing.T) {
		matches(t, true, parsed(t, "denied:\n  - git\n").Denied, "git status")
	})

	t.Run("a pattern is never read as one", func(t *testing.T) {
		denied := parsed(t, "denied:\n  - patterns: ['git push']\n").Denied
		matches(t, true, denied, "grep -rn 'git push' docs")
	})
}

func TestAnEntryWithAReason(t *testing.T) {
	t.Run("one command may be written alone", func(t *testing.T) {
		denied := parsed(t, "denied:\n  - commands: shutdown\n    reason: Ask first.\n").Denied
		matches(t, true, denied, "shutdown now")
		notes(t, denied, "shutdown now", "Ask first.")
	})

	t.Run("several share it", func(t *testing.T) {
		denied := parsed(t, "denied:\n  - commands: [shutdown, reboot]\n    reason: Ask first.\n").Denied
		notes(t, denied, "shutdown now", "Ask first.")
		notes(t, denied, "reboot now", "Ask first.")
	})

	t.Run("it is optional", func(t *testing.T) {
		notes(t, parsed(t, "denied:\n  - commands: [shutdown]\n").Denied, "shutdown now", "")
	})

	t.Run("a bare command is the same entry without one", func(t *testing.T) {
		bare := parsed(t, "denied:\n  - shutdown\n").Denied
		grouped := parsed(t, "denied:\n  - commands: shutdown\n").Denied
		if bare.Matches("shutdown now") != grouped.Matches("shutdown now") {
			t.Error("a bare command is not the same as one written out")
		}
		notes(t, bare, "shutdown now", "")
	})

	t.Run("a git command carries it too", func(t *testing.T) {
		denied := parsed(t, "denied:\n  - commands: git push\n    reason: Open a pull request.\n").Denied
		notes(t, denied, "git push", "Open a pull request.")
	})
}

func TestDescriptions(t *testing.T) {
	t.Run("a description is what the rules call themselves", func(t *testing.T) {
		denied := parsed(t, "denied:\n  - patterns: ['\\bnever\\b']\n    description: Something unreadable\n").Denied
		named(t, denied, "never", "Something unreadable")
	})

	t.Run("without one a rule calls itself what it is written as", func(t *testing.T) {
		named(t, parsed(t, "denied:\n  - patterns: ['\\bnever\\b']\n").Denied, "never", `\bnever\b`)
	})

	t.Run("it covers every command and pattern in its entry", func(t *testing.T) {
		denied := parsed(t, "denied:\n  - commands: [shutdown]\n    patterns: ['\\breboot\\b']\n    description: Downtime\n").Denied
		named(t, denied, "shutdown now", "Downtime")
		named(t, denied, "reboot now", "Downtime")
	})
}

func TestTrustedDirectories(t *testing.T) {
	const written = "trusted_git_directories: [/trusted]\ndenied: [git push]\nallowed: [git status]\n"

	t.Run("bind an allowed git command", func(t *testing.T) {
		allowed := parsed(t, written).Allowed
		matches(t, true, allowed, "git -C /trusted/repo status")
		matches(t, false, allowed, "git -C /elsewhere status")
	})

	t.Run("do not bind a denied one", func(t *testing.T) {
		matches(t, true, parsed(t, written).Denied, "git -C /elsewhere push")
	})

	t.Run("naming none leaves every -C unallowed", func(t *testing.T) {
		allowed := parsed(t, "allowed: [git status]\n").Allowed
		matches(t, true, allowed, "git status")
		matches(t, false, allowed, "git -C /anywhere status")
	})
}

func TestWhatIsRefused(t *testing.T) {
	t.Run("an entry with neither commands nor patterns", func(t *testing.T) {
		refuses(t, "denied:\n  - reason: advice with nothing to advise on\n", "commands` or `patterns")
	})

	t.Run("a pattern that is not text", func(t *testing.T) {
		refuses(t, "denied:\n  - patterns: [7]\n", "")
	})

	t.Run("an expression that will not compile", func(t *testing.T) {
		refuses(t, "denied:\n  - patterns: ['([']\n", "error parsing regexp")
		refuses(t, "allowed:\n  - patterns: ['([']\n", "error parsing regexp")
	})

	t.Run("an empty command", func(t *testing.T) {
		refuses(t, "allowed:\n  - '  '\n", "written as text")
	})
}

func TestAnEmptyConfiguration(t *testing.T) {
	t.Run("decides nothing", func(t *testing.T) {
		empty := parsed(t, "")
		matches(t, false, empty.Denied, "anything")
		matches(t, false, empty.Allowed, "anything")
	})
}

func TestReadingAFile(t *testing.T) {
	t.Run("reads a policy off disk", func(t *testing.T) {
		matches(t, true, read(t, "denied:\n  - shutdown\n").Denied, "shutdown now")
	})

	t.Run("an empty file decides nothing", func(t *testing.T) {
		matches(t, false, read(t, "").Denied, "anything")
	})

	t.Run("a file that is not there is an error", func(t *testing.T) {
		if _, err := policy.Read(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
			t.Error("wanted an error")
		}
	})
}

func TestTheBuiltInPolicy(t *testing.T) {
	t.Run("is the one this module ships", func(t *testing.T) {
		matches(t, true, policy.Default().Denied, "rm -rf /")
		matches(t, true, policy.Default().Allowed, "git status")
	})

	t.Run("gives way to a policy file that is named", func(t *testing.T) {
		chosen, err := policy.Chosen(written(t, "denied:\n  - shutdown\n"))
		if err != nil {
			t.Fatal(err)
		}
		matches(t, true, chosen.Denied, "shutdown now")
		matches(t, false, chosen.Denied, "rm -rf /")
	})
}

func TestWhichPolicyAnswers(t *testing.T) {
	t.Run("the environment gives way to what is named", func(t *testing.T) {
		t.Setenv(policy.Variable, written(t, "denied:\n  - reboot\n"))
		chosen := asked(t, written(t, "denied:\n  - shutdown\n"))
		matches(t, true, chosen.Denied, "shutdown now")
		matches(t, false, chosen.Denied, "reboot now")
	})

	t.Run("the environment answers when nothing is named", func(t *testing.T) {
		t.Setenv(policy.Variable, written(t, "denied:\n  - reboot\n"))
		matches(t, true, asked(t, "").Denied, "reboot now")
	})

	t.Run("nothing named and nothing in the environment leaves the built-in one", func(t *testing.T) {
		t.Setenv(policy.Variable, "")
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())
		matches(t, true, asked(t, "").Denied, "rm -rf /")
	})

	t.Run("a policy that answers is the only one read", func(t *testing.T) {
		t.Setenv(policy.Variable, "")
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())
		chosen := asked(t, written(t, "denied:\n  - shutdown\n"))
		matches(t, false, chosen.Denied, "rm -rf /")
		matches(t, false, chosen.Allowed, "git status")
	})

	t.Run("one that is named and not there is an error, not a fallback", func(t *testing.T) {
		if _, err := policy.Chosen(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
			t.Error("wanted an error")
		}
	})
}

func asked(t *testing.T, named string) policy.Policy {
	t.Helper()
	chosen, err := policy.Chosen(named)
	if err != nil {
		t.Fatal(err)
	}
	return chosen
}

func parsed(t *testing.T, configuration string) policy.Policy {
	t.Helper()
	built, err := policy.Parse([]byte(configuration))
	if err != nil {
		t.Fatalf("Parse(%q): %v", configuration, err)
	}
	return built
}

func read(t *testing.T, configuration string) policy.Policy {
	t.Helper()
	built, err := policy.Read(written(t, configuration))
	if err != nil {
		t.Fatalf("Read(%q): %v", configuration, err)
	}
	return built
}

func written(t *testing.T, configuration string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "policy.yaml")
	if err := os.WriteFile(path, []byte(configuration), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func refuses(t *testing.T, configuration, complaint string) {
	t.Helper()
	_, err := policy.Parse([]byte(configuration))
	if err == nil {
		t.Fatalf("Parse(%q) was accepted", configuration)
	}
	if !strings.Contains(err.Error(), complaint) {
		t.Errorf("Parse(%q) said %q, wanted it to mention %q", configuration, err, complaint)
	}
}

func matches(t *testing.T, wanted bool, group rules.Group, command string) {
	t.Helper()
	if got := group.Matches(command); got != wanted {
		t.Errorf("%s matching %q = %v, wanted %v", group, command, got, wanted)
	}
}

func notes(t *testing.T, group rules.Group, command, wanted string) {
	t.Helper()
	if got := matching(t, group, command).Note(); got != wanted {
		t.Errorf("%q -> note %q, wanted %q", command, got, wanted)
	}
}

func named(t *testing.T, group rules.Group, command, wanted string) {
	t.Helper()
	if got := matching(t, group, command).String(); got != wanted {
		t.Errorf("%q -> %q, wanted %q", command, got, wanted)
	}
}

func matching(t *testing.T, group rules.Group, command string) rules.Rule {
	t.Helper()
	matched := group.Matching(command)
	if len(matched) == 0 {
		t.Fatalf("%q matched nothing", command)
	}
	return matched[0]
}
