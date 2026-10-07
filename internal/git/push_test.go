// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package git_test

import (
	"os/exec"
	"testing"

	"github.com/thehale/tooluse-screener/internal/git"
)

func TestTheBranchAPushLandsOn(t *testing.T) {
	t.Run("a remote and a branch", func(t *testing.T) {
		lands(t, "git push origin topic", "topic")
		lands(t, "git push upstream feature/a-thing", "feature/a-thing")
	})

	t.Run("a refspec says where it lands, not where HEAD is", func(t *testing.T) {
		lands(t, "git push origin HEAD:topic", "topic")
		lands(t, "git push origin HEAD:main", "main")
		lands(t, "git push origin topic:main", "main")
	})

	t.Run("refs/heads is the same branch written out", func(t *testing.T) {
		lands(t, "git push origin HEAD:refs/heads/main", "main")
	})

	t.Run("a -C spelling lands the same way", func(t *testing.T) {
		lands(t, "git -C /elsewhere push origin topic", "topic")
	})

	t.Run("options that change neither end", func(t *testing.T) {
		lands(t, "git push -u origin topic", "topic")
		lands(t, "git push --set-upstream --quiet origin topic", "topic")
	})

	t.Run("a forced push lands where it would without force", func(t *testing.T) {
		lands(t, "git push --force origin topic", "topic")
		lands(t, "git push -f origin topic", "topic")
		lands(t, "git push --force-with-lease origin topic", "topic")
		lands(t, "git push --force-with-lease=topic:abc123 --force-if-includes origin topic", "topic")
		lands(t, "git push origin +topic:main", "main")
		lands(t, "git push origin +topic", "topic")
		landsFrom(t, []string{repository(t, "topic")}, "git push --force", "topic")
		landsFrom(t, []string{repository(t, "topic")}, "git push origin +HEAD", "topic")
	})
}

func TestWhereAPushIsUnread(t *testing.T) {
	t.Run("a spelling that does not say where it lands", func(t *testing.T) {
		isUnread(t, "git push")
		isUnread(t, "git push origin")
		isUnread(t, "git push origin topic other")
	})

	t.Run("an option that moves what lands", func(t *testing.T) {
		isUnread(t, "git push --delete origin topic")
		isUnread(t, "git push --mirror origin topic")
		isUnread(t, "git push --all origin topic")
		isUnread(t, "git push --repo=elsewhere origin topic")
		isUnread(t, "git push --receive-pack=evil origin topic")
		isUnread(t, "git push --no-verify origin topic")
	})

	t.Run("a refspec that does anything but update a branch", func(t *testing.T) {
		isUnread(t, "git push origin :topic")
		isUnread(t, "git push origin topic:refs/tags/v1")
		isUnread(t, "git push origin 'refs/heads/*:refs/heads/*'")
		isUnread(t, "git push origin HEAD:../escape")
	})

	t.Run("a global option that changes what the words mean", func(t *testing.T) {
		isUnread(t, "git -c remote.origin.pushurl=https://example.com/x.git push origin topic")
		isUnread(t, "git --namespace=sneaky push origin topic")
		isUnread(t, "git --git-dir=/elsewhere/.git push origin topic")
		isUnread(t, "git --work-tree=/elsewhere push origin topic")
		isUnread(t, "git --no-pager push origin topic")
	})

	t.Run("an assignment git reads as one of those options", func(t *testing.T) {
		isUnread(t, "GIT_DIR=/elsewhere/.git git push origin topic")
		isUnread(t, "GIT_WORK_TREE=/elsewhere git push origin topic")
		isUnread(t, "GIT_NAMESPACE=sneaky git push origin topic")
	})

	t.Run("an assignment that changes the config git reads", func(t *testing.T) {
		isUnread(t, "GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0=remote.origin.pushurl GIT_CONFIG_VALUE_0=/x git push origin topic")
		isUnread(t, "HOME=/elsewhere git push origin topic")
	})

	t.Run("a remote that is a place rather than a name", func(t *testing.T) {
		isUnread(t, "git push git@github.com:acme/repo.git topic")
		isUnread(t, "git push https://example.com/repo topic")
		isUnread(t, "git push ../another/repo topic")
	})

	t.Run("another subcommand, or none", func(t *testing.T) {
		isUnread(t, "git status")
		isUnread(t, "git pushy origin topic")
		isUnread(t, "ls origin topic")
	})
}

func TestTheBranchARepositorySays(t *testing.T) {
	t.Run("a push that names no branch lands on the one checked out", func(t *testing.T) {
		on := repository(t, "topic")
		landsFrom(t, []string{on}, "git push", "topic")
		landsFrom(t, []string{on}, "git push origin", "topic")
		landsFrom(t, []string{on}, "git push -u origin", "topic")
	})

	t.Run("HEAD is the branch checked out", func(t *testing.T) {
		landsFrom(t, []string{repository(t, "topic")}, "git push origin HEAD", "topic")
	})

	t.Run("push.default upstream lands on the branch it tracks", func(t *testing.T) {
		tracking := repository(t, "topic", "push.default", "upstream", "branch.topic.remote", "origin", "branch.topic.merge", "refs/heads/main")
		landsFrom(t, []string{tracking}, "git push", "main")
	})

	t.Run("push.default simple and current land on the same name", func(t *testing.T) {
		landsFrom(t, []string{repository(t, "topic", "push.default", "current")}, "git push", "topic")
		landsFrom(t, []string{repository(t, "topic", "push.default", "simple")}, "git push", "topic")
	})

	t.Run("a push the repository spreads over several branches", func(t *testing.T) {
		isUnreadFrom(t, []string{repository(t, "topic", "push.default", "matching")}, "git push")
		isUnreadFrom(t, []string{repository(t, "topic", "remote.origin.push", "refs/heads/*:refs/heads/*")}, "git push")
	})

	t.Run("a remote the repository names that is a place rather than a name", func(t *testing.T) {
		isUnreadFrom(t, []string{repository(t, "topic", "remote.pushDefault", "../elsewhere")}, "git push")
	})

	t.Run("directories that disagree, or one nobody can read", func(t *testing.T) {
		isUnreadFrom(t, []string{repository(t, "topic"), repository(t, "other")}, "git push")
		landsFrom(t, []string{repository(t, "topic"), repository(t, "topic")}, "git push", "topic")
		isUnreadFrom(t, []string{repository(t, "topic"), ""}, "git push")
	})

	t.Run("a directory that is no repository", func(t *testing.T) {
		isUnreadFrom(t, []string{t.TempDir()}, "git push origin HEAD")
	})
}

func repository(t *testing.T, branch string, settings ...string) string {
	t.Helper()
	dir := t.TempDir()
	runs(t, dir, "init", "--quiet", "--initial-branch", branch)
	for len(settings) > 1 {
		runs(t, dir, "config", settings[0], settings[1])
		settings = settings[2:]
	}
	return dir
}

func runs(t *testing.T, dir string, arguments ...string) {
	t.Helper()
	if said, err := exec.Command("git", append([]string{"-C", dir}, arguments...)...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", arguments, err, said)
	}
}

func lands(t *testing.T, command, wanted string) {
	t.Helper()
	landsFrom(t, []string{t.TempDir()}, command, wanted)
}

func landsFrom(t *testing.T, dirs []string, command, wanted string) {
	t.Helper()
	branch, known := git.Read(command).Landing(dirs)
	if !known || branch != wanted {
		t.Errorf("Landing(%q) = %q, %v, wanted %q, true", command, branch, known, wanted)
	}
}

func isUnread(t *testing.T, command string) {
	t.Helper()
	isUnreadFrom(t, []string{t.TempDir()}, command)
}

func isUnreadFrom(t *testing.T, dirs []string, command string) {
	t.Helper()
	if branch, known := git.Read(command).Landing(dirs); known {
		t.Errorf("Landing(%q) = %q, true, wanted it unread", command, branch)
	}
}
