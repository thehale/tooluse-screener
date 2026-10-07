// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"testing"
)

const guarding = `
denied:
  - name: Write a secret
    reason: Ask a human first.
    paths: /secrets/**
  - commands: cat /secrets
  - paths: /notes/private/*
allowed:
  - name: Take notes
    paths: /notes/**
`

func TestWritingAPath(t *testing.T) {
	guarded := parsed(t, guarding)

	t.Run("a denied path is refused with its name and reason", func(t *testing.T) {
		wrote(t, guarded, "/secrets/key", "deny: Path matches a denied rule: Write a secret. Ask a human first.")
	})

	t.Run("an allowed path is allowed", func(t *testing.T) {
		wrote(t, guarded, "/notes/today.md", "allow: Path is allowed: Take notes")
	})

	t.Run("a path both refuse and allow is refused", func(t *testing.T) {
		wrote(t, guarded, "/notes/private/diary.md", "deny: Path matches a denied rule: /notes/private/*")
	})

	t.Run("a path the policy does not name is asked about", func(t *testing.T) {
		wrote(t, guarded, "/work/main.go", "ask: Path is not in the shared allow list")
	})

	t.Run("a path is never judged by a command", func(t *testing.T) {
		wrote(t, guarded, "cat /secrets", "ask: Path is not in the shared allow list")
	})

	t.Run("a command is never judged by a path", func(t *testing.T) {
		answers(t, Ask, guarded, "/notes/today.md")
	})
}

func wrote(t *testing.T, chosen Policy, path, wanted string) {
	t.Helper()
	if verdict := chosen.CheckPath(path); verdict.String() != wanted {
		t.Errorf("%q -> %s, wanted %s", path, verdict, wanted)
	}
}
