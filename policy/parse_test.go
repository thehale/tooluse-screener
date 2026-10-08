// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thehale/tooluse-screener/internal/command"
	"github.com/thehale/tooluse-screener/internal/rules"
)

func TestWhereACommandIsLookedFor(t *testing.T) {
	t.Run("a denied command is looked for anywhere", func(t *testing.T) {
		denyRules := policyFrom(t, "denied:\n  - deploy now\n").denied.commandRules
		matches(t, true, denyRules, "deploy now")
		matches(t, true, denyRules, "RELEASE_CHANNEL=live deploy now")
	})

	t.Run("an allowed command only at the start", func(t *testing.T) {
		allowRules := policyFrom(t, "allowed:\n  - ls\n").allowed.commandRules
		matches(t, true, allowRules, "ls -la")
		matches(t, false, allowRules, "lsblk")
		matches(t, false, allowRules, "echo ls")
	})

	t.Run("a command is literal text", func(t *testing.T) {
		denyRules := policyFrom(t, "denied:\n  - a.c\n").denied.commandRules
		matches(t, true, denyRules, "a.c")
		matches(t, false, denyRules, "abc")
	})
}

func TestWhereAPatternIsLookedFor(t *testing.T) {
	t.Run("a denied pattern is looked for anywhere", func(t *testing.T) {
		denyRules := policyFrom(t, "denied:\n  - patterns: ['\\bnever\\b']\n").denied.commandRules
		matches(t, true, denyRules, "do never run this")
	})

	t.Run("an allowed pattern only at the start", func(t *testing.T) {
		allowRules := policyFrom(t, "allowed:\n  - patterns: ['ls|cat']\n").allowed.commandRules
		matches(t, true, allowRules, "cat foo")
		matches(t, false, allowRules, "echo cat")
	})
}

func TestACommandStartingWithGit(t *testing.T) {
	t.Run("is read as a git command", func(t *testing.T) {
		denyRules := policyFrom(t, "denied:\n  - git push\n").denied.commandRules
		matches(t, true, denyRules, "git push origin main")
		matches(t, true, denyRules, "git -C /anywhere push")
		matches(t, false, denyRules, "git status")
	})

	t.Run("carries its arguments", func(t *testing.T) {
		denyRules := policyFrom(t, "denied:\n  - git config --unset\n").denied.commandRules
		matches(t, true, denyRules, "git config --unset user.name")
		matches(t, true, denyRules, "git -C /x config --global --unset user.name")
		matches(t, false, denyRules, "git config --get user.name")
	})

	t.Run("is not looked for as text", func(t *testing.T) {
		denyRules := policyFrom(t, "denied:\n  - git push\n").denied.commandRules
		matches(t, false, denyRules, "grep -rn 'git push' docs")
	})

	t.Run("a bare git is left as text", func(t *testing.T) {
		matches(t, true, policyFrom(t, "denied:\n  - git\n").denied.commandRules, "git status")
	})

	t.Run("a pattern is never read as one", func(t *testing.T) {
		denyRules := policyFrom(t, "denied:\n  - patterns: ['git push']\n").denied.commandRules
		matches(t, true, denyRules, "grep -rn 'git push' docs")
	})
}

func TestAnEntryWithAReason(t *testing.T) {
	t.Run("one command may be written alone", func(t *testing.T) {
		denyRules := policyFrom(t, "denied:\n  - commands: shutdown\n    reason: Ask first.\n").denied.commandRules
		matches(t, true, denyRules, "shutdown now")
		reasons(t, denyRules, "shutdown now", "Ask first.")
	})

	t.Run("several share it", func(t *testing.T) {
		denyRules := policyFrom(t, "denied:\n  - commands: [shutdown, reboot]\n    reason: Ask first.\n").denied.commandRules
		reasons(t, denyRules, "shutdown now", "Ask first.")
		reasons(t, denyRules, "reboot now", "Ask first.")
	})

	t.Run("it is optional", func(t *testing.T) {
		reasons(t, policyFrom(t, "denied:\n  - commands: [shutdown]\n").denied.commandRules, "shutdown now", "")
	})

	t.Run("a bare command is the same entry without one", func(t *testing.T) {
		scalarRules := policyFrom(t, "denied:\n  - shutdown\n").denied.commandRules
		mappingRules := policyFrom(t, "denied:\n  - commands: shutdown\n").denied.commandRules
		if scalarRules.HasMatchFor(commandOf("shutdown now")) != mappingRules.HasMatchFor(commandOf("shutdown now")) {
			t.Error("a bare command is not the same as one written out")
		}
		reasons(t, scalarRules, "shutdown now", "")
	})

	t.Run("a git command carries it too", func(t *testing.T) {
		denyRules := policyFrom(t, "denied:\n  - commands: git push\n    reason: Open a pull request.\n").denied.commandRules
		reasons(t, denyRules, "git push", "Open a pull request.")
	})
}

func TestNames(t *testing.T) {
	t.Run("a name is what the rules call themselves", func(t *testing.T) {
		denyRules := policyFrom(t, "denied:\n  - patterns: ['\\bnever\\b']\n    name: Say never\n").denied.commandRules
		isNamed(t, denyRules, "never", "Say never")
	})

	t.Run("without one a rule calls itself what it is written as", func(t *testing.T) {
		isNamed(t, policyFrom(t, "denied:\n  - patterns: ['\\bnever\\b']\n").denied.commandRules, "never", `\bnever\b`)
	})

	t.Run("it covers every command and pattern in its entry", func(t *testing.T) {
		denyRules := policyFrom(t, "denied:\n  - commands: [shutdown]\n    patterns: ['\\breboot\\b']\n    name: Take the machine down\n").denied.commandRules
		isNamed(t, denyRules, "shutdown now", "Take the machine down")
		isNamed(t, denyRules, "reboot now", "Take the machine down")
	})

	t.Run("the description it was once written with is refused", func(t *testing.T) {
		refuses(t, "denied:\n  - commands: shutdown\n    description: Downtime\n", "Rename it to `name`")
	})
}

func TestTrustedDirectories(t *testing.T) {
	const policyYAML = "trusted_git_directories: [/trusted]\ndenied: [git push]\nallowed: [git status]\n"

	t.Run("bind an allowed git command", func(t *testing.T) {
		allowRules := policyFrom(t, policyYAML).allowed.commandRules
		matches(t, true, allowRules, "git -C /trusted/repo status")
		matches(t, false, allowRules, "git -C /elsewhere status")
	})

	t.Run("do not bind a denied one", func(t *testing.T) {
		matches(t, true, policyFrom(t, policyYAML).denied.commandRules, "git -C /elsewhere push")
	})

	t.Run("naming none leaves every -C unallowed", func(t *testing.T) {
		allowRules := policyFrom(t, "allowed: [git status]\n").allowed.commandRules
		matches(t, true, allowRules, "git status")
		matches(t, false, allowRules, "git -C /anywhere status")
	})
}

func TestOnlyDirs(t *testing.T) {
	root := t.TempDir()
	here, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("a command with no directory of its own is judged where it runs", func(t *testing.T) {
		matches(t, true, policyFrom(t, "allowed:\n  - commands: ls\n    only:\n      dirs: ["+here+"]\n").allowed.commandRules, "ls -la")
		matches(t, false, policyFrom(t, "allowed:\n  - commands: ls\n    only:\n      dirs: ["+root+"]\n").allowed.commandRules, "ls -la")
	})

	t.Run("a git command is judged by the directory it names", func(t *testing.T) {
		allowRules := policyFrom(t, "trusted_git_directories: ["+root+"]\nallowed:\n  - patterns: ['^git -C \\S+ status$']\n    only:\n      dirs: ["+root+"]\n").allowed.commandRules
		matches(t, true, allowRules, "git -C "+root+" status")
		matches(t, false, allowRules, "git -C /elsewhere status")
	})

	t.Run("scopes a denial the same way", func(t *testing.T) {
		denyRules := policyFrom(t, "denied:\n  - commands: git status\n    only:\n      dirs: ["+root+"]\n").denied.commandRules
		matches(t, true, denyRules, "git -C "+root+" status")
		matches(t, false, denyRules, "git -C /elsewhere status")
	})

	t.Run("an allowed git command still answers to trusted_git_directories as well", func(t *testing.T) {
		policyYAML := "allowed:\n  - commands: git status\n    only:\n      dirs: [" + root + "]\n"
		matches(t, false, policyFrom(t, policyYAML).allowed.commandRules, "git -C "+root+" status")
		matches(t, true, policyFrom(t, "trusted_git_directories: ["+root+"]\n"+policyYAML).allowed.commandRules, "git -C "+root+" status")
	})
}

func TestOnlyBranches(t *testing.T) {
	root := t.TempDir()
	policyYAML := "trusted_git_directories: [" + root + "]\n" +
		"denied:\n  - name: Push\n    commands: git push\n" +
		"allowed:\n  - name: Push where CI can run it\n" +
		"    patterns: ['^git (-C \\S+ )?push origin \\S+$']\n" +
		"    only:\n      dirs: [" + root + "]\n      branches: [{not: [main, master, trunk]}]\n"

	t.Run("lets a branch outside the list through", func(t *testing.T) {
		allowRules := policyFrom(t, policyYAML).allowed.commandRules
		matches(t, true, allowRules, "git -C "+root+" push origin topic")
		matches(t, true, allowRules, "git -C "+root+" push origin HEAD:topic")
	})

	t.Run("leaves out every spelling of one inside it", func(t *testing.T) {
		allowRules := policyFrom(t, policyYAML).allowed.commandRules
		matches(t, false, allowRules, "git -C "+root+" push origin main")
		matches(t, false, allowRules, "git -C "+root+" push origin HEAD:main")
		matches(t, false, allowRules, "git -C "+root+" push origin refs/heads/trunk")
		matches(t, false, allowRules, "git -C "+root+" push origin MASTER")
	})

	t.Run("leaves out a landing it cannot read at all", func(t *testing.T) {
		allowRules := policyFrom(t, policyYAML).allowed.commandRules
		matches(t, false, allowRules, "git -C "+root+" push origin +topic:main")
		matches(t, false, allowRules, "git -C /elsewhere push origin topic")
	})

	t.Run("the command is named by the rule, never implied by the scope", func(t *testing.T) {
		refuses(t, "allowed:\n  - name: Match nothing\n    only:\n      dirs: [/approved]\n", "`commands`, `patterns` or `paths`")
	})

	t.Run("one name may be written without a list", func(t *testing.T) {
		allowRules := policyFrom(t, "allowed:\n  - patterns: ['^git push origin \\S+$']\n    only:\n      branches: [{not: main}]\n").allowed.commandRules
		matches(t, true, allowRules, "git push origin topic")
		matches(t, false, allowRules, "git push origin main")
	})

	t.Run("several conditions all have to be met", func(t *testing.T) {
		allowRules := policyFrom(t, "allowed:\n  - patterns: ['^git push origin \\S+$']\n    only:\n      branches: [{not: main}, {not: [release]}]\n").allowed.commandRules
		matches(t, true, allowRules, "git push origin topic")
		matches(t, false, allowRules, "git push origin main")
		matches(t, false, allowRules, "git push origin release")
	})

	t.Run("a branch named plainly is one the push may land on", func(t *testing.T) {
		allowRules := policyFrom(t, "allowed:\n  - patterns: ['^git push origin \\S+$']\n    only:\n      branches: [ci-fix, release]\n").allowed.commandRules
		matches(t, true, allowRules, "git push origin ci-fix")
		matches(t, true, allowRules, "git push origin release")
		matches(t, false, allowRules, "git push origin main")
		matches(t, false, allowRules, "git push origin anything-else")
	})

	t.Run("names and exclusions both have to be met", func(t *testing.T) {
		allowRules := policyFrom(t, "allowed:\n  - patterns: ['^git push origin \\S+$']\n    only:\n      branches: [ci-fix, {not: ci-fix}]\n").allowed.commandRules
		matches(t, false, allowRules, "git push origin ci-fix")
	})

	t.Run("a condition saying nothing is refused", func(t *testing.T) {
		refuses(t, "allowed:\n  - commands: git push\n    only:\n      branches: [{}]\n", "branch missing `not`")
	})
}

func TestAddendum(t *testing.T) {
	root := t.TempDir()
	policyYAML := "denied:\n  - commands: git push\n    reason: Ask first.\n" +
		"    addendum:\n      - only: {dirs: [" + root + "]}\n        reason: Push your own branch by name.\n"

	t.Run("adds to the reason where its scope includes the command", func(t *testing.T) {
		reasons(t, policyFrom(t, policyYAML).denied.commandRules, "git -C "+root+" push", "Ask first. Push your own branch by name.")
	})

	t.Run("leaves the reason alone elsewhere", func(t *testing.T) {
		reasons(t, policyFrom(t, policyYAML).denied.commandRules, "git -C /elsewhere push", "Ask first.")
	})

	t.Run("needs both an only and a reason", func(t *testing.T) {
		refuses(t, "denied:\n  - commands: git push\n    addendum:\n      - reason: Anywhere.\n", "addendum missing `only`")
		refuses(t, "denied:\n  - commands: git push\n    addendum:\n      - only: {dirs: [/work]}\n", "addendum missing `reason`")
	})

	t.Run("may stand without a reason of its own", func(t *testing.T) {
		addendumOnlyYAML := "denied:\n  - commands: git push\n    addendum:\n      - only: {dirs: [" + root + "]}\n        reason: Push your own branch by name.\n"
		reasons(t, policyFrom(t, addendumOnlyYAML).denied.commandRules, "git -C "+root+" push", "Push your own branch by name.")
	})

	t.Run("adds to an allowed entry's reason the same way", func(t *testing.T) {
		here, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		allowYAML := "allowed:\n  - commands: ls\n    reason: Listing is harmless.\n" +
			"    addendum:\n      - only: {dirs: [" + here + "]}\n        reason: Even here.\n"
		reasons(t, policyFrom(t, allowYAML).allowed.commandRules, "ls", "Listing is harmless. Even here.")
	})

	t.Run("is refused beside paths", func(t *testing.T) {
		refuses(t, "denied:\n  - paths: /secrets/**\n    addendum:\n      - only: {dirs: [/work]}\n        reason: Here.\n", "`addendum` does not apply to `paths`")
	})
}

func TestPaths(t *testing.T) {
	t.Run("are read under denied and allowed alike", func(t *testing.T) {
		both := policyFrom(t, "denied:\n  - paths: /secrets/**\nallowed:\n  - paths: ['/notes/*', '/drafts/*']\n")
		namesGlobs(t, both.denied.pathRules, "/secrets/**")
		namesGlobs(t, both.allowed.pathRules, "/notes/*", "/drafts/*")
	})

	t.Run("carry their entry's reason and name", func(t *testing.T) {
		denyRules := policyFrom(t, "denied:\n  - name: Write a secret\n    reason: Ask first.\n    paths: /secrets/**\n").denied.pathRules
		namesGlobs(t, denyRules, "Write a secret")
		if reason := denyRules[0].Reason; reason != "Ask first." {
			t.Errorf("reason %q", reason)
		}
	})

	t.Run("are never read as commands", func(t *testing.T) {
		matches(t, false, policyFrom(t, "denied:\n  - paths: /secrets/**\n").denied.commandRules, "cat /secrets/key")
	})

	t.Run("one from a home other than your own is refused", func(t *testing.T) {
		refuses(t, "denied:\n  - paths: '~nobody/.ssh/**'\n", "names another user's home")
	})

	t.Run("one sharing an entry with commands or patterns is refused", func(t *testing.T) {
		refuses(t, "denied:\n  - commands: cat /secrets\n    paths: /secrets/**\n", "`paths` and `commands` in one rule")
		refuses(t, "denied:\n  - patterns: ['^cat ']\n    paths: /secrets/**\n", "`paths` and `patterns` in one rule")
	})

	t.Run("one scoped by only is refused", func(t *testing.T) {
		refuses(t, "denied:\n  - paths: /secrets/**\n    only:\n      dirs: [/work]\n", "`only` does not apply to `paths`")
	})
}

func TestWhatIsRefused(t *testing.T) {
	t.Run("an entry with neither commands nor patterns", func(t *testing.T) {
		refuses(t, "denied:\n  - reason: advice with nothing to advise on\n", "`commands`, `patterns` or `paths`")
	})

	t.Run("a pattern that is not text", func(t *testing.T) {
		refuses(t, "denied:\n  - patterns: [7]\n", "number 7 is not text")
	})

	t.Run("a value of the wrong shape names what it is and what to use", func(t *testing.T) {
		refuses(t, "- ls\n", "policy is a list. Use a mapping")
		refuses(t, "denied: ls\n", "`denied` is text `ls`. Use a list of rules.")
		refuses(t, "denied:\n  - commands: ls\n    only: 5\n", "`only` is number 5.")
		refuses(t, "denied:\n  - commands: ls\n    only:\n      branches: main\n", "`branches` is text `main`.")
	})

	t.Run("a misspelt key names the key it is closest to", func(t *testing.T) {
		refuses(t, "denid:\n  - ls\n", "`denid` is not supported. Did you mean `denied`?")
		refuses(t, "denied:\n  - commands: ls\n    only:\n      dirz: [/x]\n", "`dirz` is not supported. Did you mean `dirs`?")
	})

	t.Run("a key like none the policy reads names the keys it does", func(t *testing.T) {
		refuses(t, "colour: red\n", "`colour` is not supported. Use one of `trusted_git_directories`, `denied` or `allowed`.")
	})

	t.Run("YAML that will not parse", func(t *testing.T) {
		refuses(t, "denied:\n  - commands: ls\n  bad indent\n", "line 3: invalid YAML")
	})

	t.Run("an expression that will not compile", func(t *testing.T) {
		refuses(t, "denied:\n  - patterns: ['([']\n", "invalid regular expression")
		refuses(t, "allowed:\n  - patterns: ['([']\n", "invalid regular expression")
	})

	t.Run("an empty command", func(t *testing.T) {
		refuses(t, "allowed:\n  - '  '\n", "blank value")
	})
}

func TestAnEmptyConfiguration(t *testing.T) {
	t.Run("decides nothing", func(t *testing.T) {
		emptyPolicy := policyFrom(t, "")
		matches(t, false, emptyPolicy.denied.commandRules, "anything")
		matches(t, false, emptyPolicy.allowed.commandRules, "anything")
	})
}

func TestReadingAFile(t *testing.T) {
	t.Run("reads a policy off disk", func(t *testing.T) {
		matches(t, true, policyInFile(t, "denied:\n  - shutdown\n").denied.commandRules, "shutdown now")
	})

	t.Run("an empty file decides nothing", func(t *testing.T) {
		matches(t, false, policyInFile(t, "").denied.commandRules, "anything")
	})

	t.Run("a file that is not there is an error", func(t *testing.T) {
		if _, err := LoadFile(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
			t.Error("wanted an error")
		}
	})

	t.Run("a refused policy names the file it came from", func(t *testing.T) {
		path := policyFile(t, "denied:\n  - commands: shutdown\n    paths: /secrets/**\n")
		_, err := LoadFile(path)
		wanted := path + ": line "
		if err == nil {
			t.Fatal("wanted an error")
		} else if !strings.HasPrefix(err.Error(), wanted) {
			t.Errorf("error %q, wanted a prefix of %q", err.Error(), wanted)
		}
	})
}

func TestTheBuiltInPolicy(t *testing.T) {
	t.Run("is the one this module ships", func(t *testing.T) {
		matches(t, true, Default().denied.commandRules, "rm -rf /")
		matches(t, true, Default().allowed.commandRules, "git status")
	})

	t.Run("gives way to a policy file that is named", func(t *testing.T) {
		filePolicy := policyFromFile(t, policyFile(t, "denied:\n  - shutdown\n"))
		matches(t, true, filePolicy.denied.commandRules, "shutdown now")
		matches(t, false, filePolicy.denied.commandRules, "rm -rf /")
	})
}

func TestWhichPolicyAnswers(t *testing.T) {
	t.Run("a file named outright pays the environment no mind", func(t *testing.T) {
		t.Setenv(Variable, policyFile(t, "denied:\n  - reboot\n"))
		filePolicy := policyFromFile(t, policyFile(t, "denied:\n  - shutdown\n"))
		matches(t, true, filePolicy.denied.commandRules, "shutdown now")
		matches(t, false, filePolicy.denied.commandRules, "reboot now")
	})

	t.Run("the environment answers first", func(t *testing.T) {
		t.Setenv(Variable, policyFile(t, "denied:\n  - reboot\n"))
		matches(t, true, defaultPolicy(t).denied.commandRules, "reboot now")
	})

	t.Run("then the config directory", func(t *testing.T) {
		t.Setenv(Variable, "")
		home := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", home)
		configPath := filepath.Join(home, "tooluse-screener", "policy.yaml")
		if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(configPath, []byte("denied:\n  - reboot\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		matches(t, true, defaultPolicy(t).denied.commandRules, "reboot now")
	})

	t.Run("and nothing in either leaves the built-in one", func(t *testing.T) {
		t.Setenv(Variable, "")
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())
		matches(t, true, defaultPolicy(t).denied.commandRules, "rm -rf /")
	})

	t.Run("a policy that answers is the only one read", func(t *testing.T) {
		filePolicy := policyFromFile(t, policyFile(t, "denied:\n  - shutdown\n"))
		matches(t, false, filePolicy.denied.commandRules, "rm -rf /")
		matches(t, false, filePolicy.allowed.commandRules, "git status")
	})

	t.Run("one that is named and not there is an error, not a fallback", func(t *testing.T) {
		if _, err := LoadFile(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
			t.Error("wanted an error")
		}
	})
}

func defaultPolicy(t *testing.T) Policy {
	t.Helper()
	policy, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

func policyFromFile(t *testing.T, path string) Policy {
	t.Helper()
	policy, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

func policyFrom(t *testing.T, configuration string) Policy {
	t.Helper()
	policy, err := Parse([]byte(configuration))
	if err != nil {
		t.Fatalf("Parse(%q): %v", configuration, err)
	}
	return policy
}

func policyInFile(t *testing.T, configuration string) Policy {
	t.Helper()
	policy, err := LoadFile(policyFile(t, configuration))
	if err != nil {
		t.Fatalf("Read(%q): %v", configuration, err)
	}
	return policy
}

func policyFile(t *testing.T, configuration string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "policy.yaml")
	if err := os.WriteFile(path, []byte(configuration), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func refuses(t *testing.T, configuration, complaint string) {
	t.Helper()
	_, err := Parse([]byte(configuration))
	if err == nil {
		t.Fatalf("Parse(%q) was accepted", configuration)
	}
	if !strings.Contains(err.Error(), complaint) {
		t.Errorf("Parse(%q) said %q, wanted it to mention %q", configuration, err, complaint)
	}
}

func matches(t *testing.T, wanted bool, group rules.Group, command string) {
	t.Helper()
	if got := group.HasMatchFor(commandOf(command)); got != wanted {
		t.Errorf("%s matching %q = %v, wanted %v", group, command, got, wanted)
	}
}

func reasons(t *testing.T, group rules.Group, command, wanted string) {
	t.Helper()
	if got := firstMatch(t, group, command).ReasonFor(commandOf(command)); got != wanted {
		t.Errorf("%q -> reason %q, wanted %q", command, got, wanted)
	}
}

func isNamed(t *testing.T, group rules.Group, command, wanted string) {
	t.Helper()
	if got := firstMatch(t, group, command).String(); got != wanted {
		t.Errorf("%q -> %q, wanted %q", command, got, wanted)
	}
}

func firstMatch(t *testing.T, group rules.Group, command string) rules.Rule {
	t.Helper()
	matches := group.RulesMatching(commandOf(command))
	if len(matches) == 0 {
		t.Fatalf("%q matched nothing", command)
	}
	return matches[0]
}

func namesGlobs(t *testing.T, globs rules.Globs, wanted ...string) {
	t.Helper()
	var names []string
	for _, glob := range globs {
		names = append(names, glob.String())
	}
	if strings.Join(names, " ") != strings.Join(wanted, " ") {
		t.Errorf("read %q, wanted %q", names, wanted)
	}
}

func commandOf(text string) command.Command {
	return command.Command{Text: text}
}
