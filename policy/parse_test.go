// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadingAFile(t *testing.T) {
	t.Run("reads a policy off disk", func(t *testing.T) {
		policy := policyFrom(t, "denies-shutdown.yaml")
		decides(t, policy, "shutdown now", Deny)
	})

	t.Run("an empty file decides nothing", func(t *testing.T) {
		policy := policyFrom(t, "empty.yaml")
		decides(t, policy, "anything", Ask)
	})

	t.Run("a file that is not there is an error", func(t *testing.T) {
		if _, err := LoadFile(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
			t.Error("wanted an error")
		}
	})

	t.Run("a refused policy names the file it came from", func(t *testing.T) {
		path := policyPath("refused.yaml")
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
		decides(t, Default(), "rm -rf /", Deny)
		decides(t, Default(), "git status", Allow)
	})

	t.Run("gives way to a policy file that is named", func(t *testing.T) {
		policy := policyFrom(t, "denies-shutdown.yaml")
		decides(t, policy, "rm -rf /", Ask)
	})
}

func TestWhichPolicyAnswers(t *testing.T) {
	t.Run("a file named outright pays the environment no mind", func(t *testing.T) {
		t.Setenv(Variable, policyPath("denies-reboot.yaml"))
		policy := policyFrom(t, "denies-shutdown.yaml")
		decides(t, policy, "reboot now", Ask)
	})

	t.Run("the environment answers first", func(t *testing.T) {
		t.Setenv(Variable, policyPath("denies-reboot.yaml"))
		decides(t, defaultPolicy(t), "reboot now", Deny)
	})

	t.Run("then the config directory", func(t *testing.T) {
		t.Setenv(Variable, "")
		t.Setenv("XDG_CONFIG_HOME", configHome(t, "denies-reboot.yaml"))
		decides(t, defaultPolicy(t), "reboot now", Deny)
	})

	t.Run("and nothing in either leaves the built-in one", func(t *testing.T) {
		t.Setenv(Variable, "")
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())
		decides(t, defaultPolicy(t), "rm -rf /", Deny)
	})

	t.Run("a policy that answers is the only one read", func(t *testing.T) {
		policy := policyFrom(t, "denies-shutdown.yaml")
		decides(t, policy, "rm -rf /", Ask)
		decides(t, policy, "git status", Ask)
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

func decides(t *testing.T, policy Policy, command string, wanted Decision) {
	t.Helper()
	if decision := policy.CheckCommand(command).Decision; decision != wanted {
		t.Errorf("%s -> %s, wanted %s", command, decision, wanted)
	}
}

func policyFrom(t *testing.T, name string) Policy {
	t.Helper()
	policy, err := LoadFile(policyPath(name))
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

func policyPath(name string) string {
	return filepath.Join("testdata", "policy-files", name)
}

func configHome(t *testing.T, name string) string {
	t.Helper()
	contents, err := os.ReadFile(policyPath(name))
	if err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	configPath := filepath.Join(home, "tooluse-screener", "policy.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	} else if err := os.WriteFile(configPath, contents, 0o600); err != nil {
		t.Fatal(err)
	}
	return home
}
