// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"os"
	"testing"
)

func TestJudgingInADirectory(t *testing.T) {
	const workdir = "/path/to/workdir"

	t.Run("a relative path is judged from the directory", func(t *testing.T) {
		policy := parsed(t, "denied:\n  - paths:\n      - 'secrets/**'\n").InDir(workdir)
		writes(t, policy, "secrets/key", Deny)
		writes(t, policy, workdir+"/secrets/key", Deny)
		writes(t, policy, "/elsewhere/secrets/key", Ask)
	})

	t.Run("a rule only for some directories sees the directory", func(t *testing.T) {
		policy := parsed(t, "allowed:\n  - commands: ls\n    only:\n      dirs:\n        - "+workdir+"\n")
		decides(t, policy.InDir(workdir), "ls", Allow)
		decides(t, policy.InDir(workdir), "cd .. && ls", Ask)
		decides(t, policy.InDir("/elsewhere"), "ls", Ask)
	})

	t.Run("a trusted git directory sees the directory", func(t *testing.T) {
		policy := parsed(t, "trusted_git_directories:\n  - "+workdir+"\nallowed:\n  - git status\n")
		decides(t, policy.InDir(workdir), "git -C repo status", Allow)
		decides(t, policy.InDir("/elsewhere"), "git -C repo status", Ask)
	})

	t.Run("the process stays where it was", func(t *testing.T) {
		before, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		parsed(t, "allowed:\n  - commands: ls\n").InDir(workdir).CheckCommand("ls")
		if after, _ := os.Getwd(); after != before {
			t.Errorf("working directory moved from %s to %s", before, after)
		}
	})
}

func parsed(t *testing.T, text string) Policy {
	t.Helper()
	policy, err := Parse([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

func writes(t *testing.T, policy Policy, path string, wanted Decision) {
	t.Helper()
	if decision := policy.CheckPath(path).Decision; decision != wanted {
		t.Errorf("%s -> %s, wanted %s", path, decision, wanted)
	}
}
