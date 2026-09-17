// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package git_test

import (
	"slices"
	"testing"

	"github.com/thehale/tooluse-screener/internal/git"
)

func TestReading(t *testing.T) {
	t.Run("a command that is not git", func(t *testing.T) {
		subcommand(t, "ls -la", "")
		subcommand(t, "github status", "")
		subcommand(t, "", "")
	})

	t.Run("a bare subcommand", func(t *testing.T) {
		subcommand(t, "git status", "status")
	})

	t.Run("a bare git has no subcommand", func(t *testing.T) {
		subcommand(t, "git", "")
	})

	t.Run("the arguments come with it", func(t *testing.T) {
		arguments(t, "git log --oneline -n 5", "--oneline", "-n", "5")
	})

	t.Run("environment assignments are passed over", func(t *testing.T) {
		subcommand(t, "GIT_TRACE=1 LANG=C git status", "status")
	})

	t.Run("something that only looks like an assignment is not", func(t *testing.T) {
		subcommand(t, "a-b=1 git status", "")
		subcommand(t, "1a=1 git status", "")
		subcommand(t, "=1 git status", "")
	})

	t.Run("an assignment a shell would not accept is not one", func(t *testing.T) {
		subcommand(t, "café=1 git status", "")
		subcommand(t, "\u0e33=1 git status", "")
	})
}

func TestGlobalOptions(t *testing.T) {
	t.Run("a flag is passed over", func(t *testing.T) {
		subcommand(t, "git --no-pager log", "log")
	})

	t.Run("a value is passed over with its option", func(t *testing.T) {
		subcommand(t, "git -c user.name=x commit -m hi", "commit")
	})

	t.Run("a joined value takes no extra word", func(t *testing.T) {
		subcommand(t, "git --git-dir=/tmp/x status", "status")
	})

	t.Run("an option with nothing after it", func(t *testing.T) {
		subcommand(t, "git -C", "")
	})
}

func TestDirectories(t *testing.T) {
	t.Run("one directory", func(t *testing.T) {
		pointed(t, "git -C /tmp/foo status", "/tmp/foo")
	})

	t.Run("a directory joined by an equals sign", func(t *testing.T) {
		pointed(t, "git -C=/tmp/foo status", "/tmp/foo")
	})

	t.Run("every directory in order", func(t *testing.T) {
		pointed(t, "git -C /a -C /b status", "/a", "/b")
	})

	t.Run("a directory after another option", func(t *testing.T) {
		pointed(t, "git --no-pager -C /tmp/foo log", "/tmp/foo")
	})

	t.Run("a second directory is not the subcommand", func(t *testing.T) {
		subcommand(t, "git -C /a -C /b status", "status")
	})

	t.Run("a command with no directories", func(t *testing.T) {
		pointed(t, "git status")
	})
}

func TestIsA(t *testing.T) {
	t.Run("the subcommand it is", func(t *testing.T) {
		if !git.Read("git status").IsA("status") {
			t.Error("git status is not a status")
		}
	})

	t.Run("not another one", func(t *testing.T) {
		if git.Read("git status").IsA("stat") {
			t.Error("git status is a stat")
		}
	})
}

func TestCarries(t *testing.T) {
	t.Run("an option counts anywhere", func(t *testing.T) {
		carries(t, true, "git config --global --unset x", "--unset")
	})

	t.Run("a missing option does not", func(t *testing.T) {
		carries(t, false, "git config --get x", "--unset")
	})

	t.Run("a bare word counts only first", func(t *testing.T) {
		carries(t, true, "git config get x", "get")
		carries(t, false, "git config alias.x get", "get")
	})

	t.Run("all of them are needed", func(t *testing.T) {
		carries(t, false, "git config --global x", "--global", "--unset")
	})

	t.Run("nothing asked for is always carried", func(t *testing.T) {
		carries(t, true, "git status")
	})
}

func subcommand(t *testing.T, command, wanted string) {
	t.Helper()
	if read := git.Read(command).Subcommand; read != wanted {
		t.Errorf("Read(%q).Subcommand = %q, wanted %q", command, read, wanted)
	}
}

func arguments(t *testing.T, command string, wanted ...string) {
	t.Helper()
	if read := git.Read(command).Arguments; !slices.Equal(read, wanted) {
		t.Errorf("Read(%q).Arguments = %q, wanted %q", command, read, wanted)
	}
}

func pointed(t *testing.T, command string, wanted ...string) {
	t.Helper()
	if read := git.Read(command).Directories; !slices.Equal(read, wanted) {
		t.Errorf("Read(%q).Directories = %q, wanted %q", command, read, wanted)
	}
}

func carries(t *testing.T, wanted bool, command string, words ...string) {
	t.Helper()
	if read := git.Read(command).Carries(words); read != wanted {
		t.Errorf("Read(%q).Carries(%q) = %v, wanted %v", command, words, read, wanted)
	}
}
