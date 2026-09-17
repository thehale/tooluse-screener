// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package git_test

import (
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
}

func TestWhereAPushIsUnread(t *testing.T) {
	t.Run("a spelling that does not say where it lands", func(t *testing.T) {
		unread(t, "git push")
		unread(t, "git push origin")
		unread(t, "git push origin topic other")
	})

	t.Run("an option that moves what lands", func(t *testing.T) {
		unread(t, "git push --force origin topic")
		unread(t, "git push -f origin topic")
		unread(t, "git push --force-with-lease origin topic")
		unread(t, "git push --delete origin topic")
		unread(t, "git push --mirror origin topic")
		unread(t, "git push --all origin topic")
		unread(t, "git push --repo=elsewhere origin topic")
		unread(t, "git push --receive-pack=evil origin topic")
		unread(t, "git push --no-verify origin topic")
	})

	t.Run("a refspec that does anything but fast-forward a branch", func(t *testing.T) {
		unread(t, "git push origin +topic:main")
		unread(t, "git push origin +topic")
		unread(t, "git push origin :topic")
		unread(t, "git push origin topic:refs/tags/v1")
		unread(t, "git push origin 'refs/heads/*:refs/heads/*'")
		unread(t, "git push origin HEAD:../escape")
	})

	t.Run("a global option that changes what the words mean", func(t *testing.T) {
		unread(t, "git -c remote.origin.pushurl=https://example.com/x.git push origin topic")
		unread(t, "git --namespace=sneaky push origin topic")
		unread(t, "git --git-dir=/elsewhere/.git push origin topic")
		unread(t, "git --work-tree=/elsewhere push origin topic")
		unread(t, "git --no-pager push origin topic")
	})

	t.Run("a remote that is a place rather than a name", func(t *testing.T) {
		unread(t, "git push git@github.com:acme/repo.git topic")
		unread(t, "git push https://example.com/repo topic")
		unread(t, "git push ../another/repo topic")
	})

	t.Run("another subcommand, or none", func(t *testing.T) {
		unread(t, "git status")
		unread(t, "git pushy origin topic")
		unread(t, "ls origin topic")
	})
}

func lands(t *testing.T, command, wanted string) {
	t.Helper()
	branch, known := git.Landing(git.Read(command))
	if !known || branch != wanted {
		t.Errorf("Landing(%q) = %q, %v, wanted %q, true", command, branch, known, wanted)
	}
}

func unread(t *testing.T, command string) {
	t.Helper()
	if branch, known := git.Landing(git.Read(command)); known {
		t.Errorf("Landing(%q) = %q, true, wanted it unread", command, branch)
	}
}
