// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package files_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thehale/tooluse-screener/internal/files"
)

func TestWithSuffix(t *testing.T) {
	t.Run("matching files are found, in sorted order", func(t *testing.T) {
		root := t.TempDir()
		touch(t, filepath.Join(root, "b.test"))
		touch(t, filepath.Join(root, "sub", "a.test"))
		found, err := files.WithSuffix(".test", root)
		if err != nil || len(found) != 2 || found[0] != filepath.Join(root, "b.test") || found[1] != filepath.Join(root, "sub", "a.test") {
			t.Errorf("WithSuffix(%q) = %q, %v", root, found, err)
		}
	})

	t.Run("a file without the suffix is ignored", func(t *testing.T) {
		root := t.TempDir()
		touch(t, filepath.Join(root, "seen.test"))
		touch(t, filepath.Join(root, "notes.md"))
		found, err := files.WithSuffix(".test", root)
		if err != nil || len(found) != 1 || found[0] != filepath.Join(root, "seen.test") {
			t.Errorf("WithSuffix(%q) = %q, %v", root, found, err)
		}
	})

	t.Run("a hidden subdirectory is skipped", func(t *testing.T) {
		root := t.TempDir()
		touch(t, filepath.Join(root, "seen.test"))
		touch(t, filepath.Join(root, ".hidden", "unseen.test"))
		found, err := files.WithSuffix(".test", root)
		if err != nil || len(found) != 1 || found[0] != filepath.Join(root, "seen.test") {
			t.Errorf("WithSuffix(%q) = %q, %v", root, found, err)
		}
	})

	t.Run("a symlinked directory named directly is walked, under its own name", func(t *testing.T) {
		root := t.TempDir()
		touch(t, filepath.Join(root, "real", "linked.test"))
		link := filepath.Join(root, "link")
		if err := os.Symlink(filepath.Join(root, "real"), link); err != nil {
			t.Fatal(err)
		}
		found, err := files.WithSuffix(".test", link)
		if err != nil || len(found) != 1 || found[0] != filepath.Join(link, "linked.test") {
			t.Errorf("WithSuffix(%q) = %q, %v", link, found, err)
		}
	})

	t.Run("a symlinked directory found along the way is not followed", func(t *testing.T) {
		root := t.TempDir()
		touch(t, filepath.Join(root, "walked", "seen.test"))
		touch(t, filepath.Join(root, "elsewhere", "unseen.test"))
		if err := os.Symlink(filepath.Join(root, "elsewhere"), filepath.Join(root, "walked", "link")); err != nil {
			t.Fatal(err)
		}
		found, err := files.WithSuffix(".test", filepath.Join(root, "walked"))
		if err != nil || len(found) != 1 || found[0] != filepath.Join(root, "walked", "seen.test") {
			t.Errorf("WithSuffix(walked) = %q, %v", found, err)
		}
	})

	t.Run("several paths are all searched", func(t *testing.T) {
		first, second := t.TempDir(), t.TempDir()
		touch(t, filepath.Join(first, "a.test"))
		touch(t, filepath.Join(second, "b.test"))
		found, err := files.WithSuffix(".test", first, second)
		if err != nil || len(found) != 2 {
			t.Errorf("WithSuffix(first, second) = %q, %v", found, err)
		}
	})

	t.Run("a missing path is reported", func(t *testing.T) {
		if _, err := files.WithSuffix(".test", filepath.Join(t.TempDir(), "missing")); err == nil {
			t.Error("WithSuffix of a missing directory succeeded")
		}
	})
}

func touch(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
}
