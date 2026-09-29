// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command_test

import (
	"os"
	"slices"
	"testing"

	"github.com/thehale/tooluse-screener/internal/command"
)

func TestMoves(t *testing.T) {
	t.Run("a line that goes nowhere moves nothing", func(t *testing.T) {
		moved(t, "ls && git status")
	})

	t.Run("a command after cd may be where cd went", func(t *testing.T) {
		moved(t, "cd /work && git status", "/work")
		moved(t, "pushd /work && git status", "/work")
		moved(t, `cd "/work" ; git status`, "/work")
		moved(t, "cd -- /work | git status", "/work")
	})

	t.Run("a command before it has not moved", func(t *testing.T) {
		if first := command.All("git status && cd /work")[0]; len(first.Moved) > 0 {
			t.Errorf("%q moved to %q", first.Text, first.Moved)
		}
	})

	t.Run("every cd is taken from everywhere the line may be", func(t *testing.T) {
		moved(t, "cd /work; cd app; git status", "/work", "/work/app", "app")
	})

	t.Run("a step back is taken the way cd takes it", func(t *testing.T) {
		moved(t, "cd /work/app/../lib && git status", "/work/lib")
		moved(t, "cd app && cd .. && git status", "..", "app")
	})

	t.Run("a bare cd goes home", func(t *testing.T) {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skip("no home directory")
		}
		moved(t, "cd && git status", home)
		moved(t, "cd ~/work && git status", home+"/work")
	})

	t.Run("an assignment that points it elsewhere, kept for the commands after it, is unknown", func(t *testing.T) {
		moved(t, "export GIT_DIR=/elsewhere/.git && git status", "")
		moved(t, "GH_REPO=other/repo; gh pr view", "")
		moved(t, "export HOME && git status", "")
	})

	t.Run("an assignment that points nowhere moves nothing", func(t *testing.T) {
		moved(t, "export GIT_TRACE=1 && git status")
		moved(t, "LANG=C; git status")
	})

	t.Run("where the shell decides is unknown", func(t *testing.T) {
		for _, line := range []string{
			`cd "$REPO" && git status`,
			"cd - && git status",
			"cd -P /work && git status",
			"cd ~nobody && git status",
			"pushd && git status",
			"pushd +1 && git status",
			"popd && git status",
			"source env.sh && git status",
			". env.sh && git status",
			`eval "$MOVE" && git status`,
		} {
			moved(t, line, "")
		}
	})
}

func moved(t *testing.T, line string, wanted ...string) {
	t.Helper()
	all := command.All(line)
	if last := all[len(all)-1]; !slices.Equal(last.Moved, wanted) {
		t.Errorf("All(%q) moved %q to %q, wanted %q", line, last.Text, last.Moved, wanted)
	}
}
