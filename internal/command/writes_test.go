// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command_test

import (
	"slices"
	"testing"

	"github.com/thehale/tooluse-screener/internal/command"
)

func TestWrites(t *testing.T) {
	t.Run("a redirect to a file is a write", func(t *testing.T) {
		writes(t, "echo hi > out.log", "out.log")
		writes(t, "echo hi >> out.log", "out.log")
		writes(t, "echo hi >| out.log", "out.log")
		writes(t, "echo hi &> out.log", "out.log")
		writes(t, "echo hi &>> out.log", "out.log")
		writes(t, "echo hi <> out.log", "out.log")
	})

	t.Run("a duplication to a file descriptor writes a file", func(t *testing.T) {
		writes(t, "echo hi >&out.log", "out.log")
	})

	t.Run("a duplication to a numbered descriptor writes nothing", func(t *testing.T) {
		writes(t, "echo hi 2>&1")
		writes(t, "echo hi >&2")
	})

	t.Run("closing a descriptor writes nothing", func(t *testing.T) {
		writes(t, "echo hi 2>&-")
	})

	t.Run("reading a file is not a write", func(t *testing.T) {
		writes(t, "cat < in.txt")
	})

	t.Run("a target the shell works out is unresolved", func(t *testing.T) {
		writes(t, `echo hi > "$out"`, "")
	})

	t.Run("a compound statement's redirect attaches to the first command inside it", func(t *testing.T) {
		all := command.All("{ echo a; echo b; } > out.log")
		if !slices.Equal(all[0].Writes, []string{"out.log"}) {
			t.Errorf("%q writes %q, wanted %q", all[0].Text, all[0].Writes, []string{"out.log"})
		}
		if all[1].Writes != nil {
			t.Errorf("%q writes %q, wanted none", all[1].Text, all[1].Writes)
		}
	})
}

func writes(t *testing.T, line string, wanted ...string) {
	t.Helper()
	var got []string
	for _, one := range command.All(line) {
		got = append(got, one.Writes...)
	}
	if !slices.Equal(got, wanted) {
		t.Errorf("All(%q) writes %q, wanted %q", line, got, wanted)
	}
}

func TestPathsWritten(t *testing.T) {
	t.Run("a write stands where the command ran", func(t *testing.T) {
		pathsWritten(t, command.Command{Writes: []string{"out.log"}}, "out.log")
	})

	t.Run("a relative write is also taken from where the line moved", func(t *testing.T) {
		pathsWritten(t, command.Command{Writes: []string{"out.log"}, Moved: []string{"/work"}},
			"out.log", "/work/out.log")
	})

	t.Run("an absolute write stands alone wherever the line moved", func(t *testing.T) {
		pathsWritten(t, command.Command{Writes: []string{"/etc/hosts"}, Moved: []string{"/work"}},
			"/etc/hosts")
	})

	t.Run("an unresolved write is unresolved wherever the line moved", func(t *testing.T) {
		pathsWritten(t, command.Command{Writes: []string{""}, Moved: []string{"/work"}}, "", "")
	})

	t.Run("an absolute write stands alone even where the line moved unknowably", func(t *testing.T) {
		pathsWritten(t, command.Command{Writes: []string{"/dev/null"}, Moved: []string{""}}, "/dev/null")
	})
}

func pathsWritten(t *testing.T, one command.Command, wanted ...string) {
	t.Helper()
	if got := one.PathsWritten(); !slices.Equal(got, wanted) {
		t.Errorf("PathsWritten() = %q, wanted %q", got, wanted)
	}
}
