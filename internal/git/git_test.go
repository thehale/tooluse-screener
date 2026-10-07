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
		hasSubcommand(t, "ls -la", "")
		hasSubcommand(t, "github status", "")
		hasSubcommand(t, "", "")
	})

	t.Run("a bare subcommand", func(t *testing.T) {
		hasSubcommand(t, "git status", "status")
	})

	t.Run("a bare git has no subcommand", func(t *testing.T) {
		hasSubcommand(t, "git", "")
	})

	t.Run("the arguments come with it", func(t *testing.T) {
		hasArguments(t, "git log --oneline -n 5", "--oneline", "-n", "5")
	})

	t.Run("environment assignments are passed over", func(t *testing.T) {
		hasSubcommand(t, "GIT_TRACE=1 LANG=C git status", "status")
	})

	t.Run("something that only looks like an assignment is not", func(t *testing.T) {
		hasSubcommand(t, "a-b=1 git status", "")
		hasSubcommand(t, "1a=1 git status", "")
		hasSubcommand(t, "=1 git status", "")
	})

	t.Run("an assignment a shell would not accept is not one", func(t *testing.T) {
		hasSubcommand(t, "café=1 git status", "")
		hasSubcommand(t, "\u0e33=1 git status", "")
	})
}

func TestGlobalOptions(t *testing.T) {
	t.Run("a flag is passed over", func(t *testing.T) {
		hasSubcommand(t, "git --no-pager log", "log")
	})

	t.Run("a value is passed over with its option", func(t *testing.T) {
		hasSubcommand(t, "git -c user.name=x commit -m hi", "commit")
	})

	t.Run("a joined value takes no extra word", func(t *testing.T) {
		hasSubcommand(t, "git --git-dir=/tmp/x status", "status")
	})

	t.Run("an option with nothing after it", func(t *testing.T) {
		hasSubcommand(t, "git -C", "")
	})
}

func TestDirectories(t *testing.T) {
	t.Run("one directory", func(t *testing.T) {
		pointsAt(t, "git -C /tmp/foo status", "/tmp/foo")
	})

	t.Run("a directory joined by an equals sign", func(t *testing.T) {
		pointsAt(t, "git -C=/tmp/foo status", "/tmp/foo")
	})

	t.Run("every directory in order", func(t *testing.T) {
		pointsAt(t, "git -C /a -C /b status", "/a", "/b")
	})

	t.Run("a directory after another option", func(t *testing.T) {
		pointsAt(t, "git --no-pager -C /tmp/foo log", "/tmp/foo")
	})

	t.Run("a second directory is not the subcommand", func(t *testing.T) {
		hasSubcommand(t, "git -C /a -C /b status", "status")
	})

	t.Run("a repository named instead of entered", func(t *testing.T) {
		pointsAt(t, "git --git-dir=/tmp/foo/.git status", "/tmp/foo/.git")
		pointsAt(t, "git --work-tree /tmp/foo status", "/tmp/foo")
		pointsAt(t, "git --git-dir=/a/.git --work-tree=/b status", "/a/.git", "/b")
	})

	t.Run("each directory entered is entered from the one before", func(t *testing.T) {
		pointsAt(t, "git -C /a -C b -C ../c status", "/a", "/a/b", "/a/b/../c")
	})

	t.Run("a repository named is named from the directory entered last", func(t *testing.T) {
		pointsAt(t, "git --git-dir=.git -C /a -C b status", "/a", "/a/b", "/a/b/.git")
		pointsAt(t, "git -C /a --work-tree=/b status", "/a", "/b")
	})

	t.Run("an assignment git reads as an option points it the same way", func(t *testing.T) {
		pointsAt(t, "GIT_DIR=/a/.git git status", "/a/.git")
		pointsAt(t, "GIT_WORK_TREE=/b git status", "/b")
		pointsAt(t, "GIT_DIR=.git git -C /a status", "/a", "/a/.git")
	})

	t.Run("an assignment git reads as nothing points it nowhere", func(t *testing.T) {
		pointsAt(t, "GIT_TRACE=1 git status")
	})

	t.Run("a directory is read as the shell passes it on", func(t *testing.T) {
		pointsAt(t, `git -C "/tmp/foo" status`, "/tmp/foo")
		pointsAt(t, `git --git-dir='/tmp/foo/.git' status`, "/tmp/foo/.git")
	})

	t.Run("a directory the shell works out is unknown", func(t *testing.T) {
		pointsAt(t, `git -C "$REPO" status`, "")
		pointsAt(t, "git -C /work/* status", "")
	})

	t.Run("a command with no directories", func(t *testing.T) {
		pointsAt(t, "git status")
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

func TestHasAll(t *testing.T) {
	t.Run("an option counts anywhere", func(t *testing.T) {
		hasAll(t, true, "git config --global --unset x", "--unset")
	})

	t.Run("a missing option does not", func(t *testing.T) {
		hasAll(t, false, "git config --get x", "--unset")
	})

	t.Run("a bare word counts only first", func(t *testing.T) {
		hasAll(t, true, "git config get x", "get")
		hasAll(t, false, "git config alias.x get", "get")
	})

	t.Run("all of them are needed", func(t *testing.T) {
		hasAll(t, false, "git config --global x", "--global", "--unset")
	})

	t.Run("nothing asked for is always carried", func(t *testing.T) {
		hasAll(t, true, "git status")
	})
}

func hasSubcommand(t *testing.T, command, wanted string) {
	t.Helper()
	if read := git.Read(command).Subcommand; read != wanted {
		t.Errorf("Read(%q).Subcommand = %q, wanted %q", command, read, wanted)
	}
}

func hasArguments(t *testing.T, command string, wanted ...string) {
	t.Helper()
	if read := git.Read(command).Arguments; !slices.Equal(read, wanted) {
		t.Errorf("Read(%q).Arguments = %q, wanted %q", command, read, wanted)
	}
}

func pointsAt(t *testing.T, command string, wanted ...string) {
	t.Helper()
	if read := git.Read(command).Directories; !slices.Equal(read, wanted) {
		t.Errorf("Read(%q).Directories = %q, wanted %q", command, read, wanted)
	}
}

func hasAll(t *testing.T, wanted bool, command string, words ...string) {
	t.Helper()
	if read := git.Read(command).HasAll(words); read != wanted {
		t.Errorf("Read(%q).HasAll(%q) = %v, wanted %v", command, words, read, wanted)
	}
}
