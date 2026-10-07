// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package directories_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thehale/tooluse-screener/internal/directories"
)

func TestAreAllUnder(t *testing.T) {
	root := t.TempDir()

	t.Run("a root contains itself", func(t *testing.T) {
		areUnder(t, true, []string{root}, []string{root})
	})

	t.Run("a root contains what is nested in it", func(t *testing.T) {
		areUnder(t, true, []string{root + "/nested/deeper"}, []string{root})
	})

	t.Run("a sibling is not inside", func(t *testing.T) {
		areUnder(t, false, []string{root + "/foobar"}, []string{root + "/foo"})
	})

	t.Run("any of the roots will do", func(t *testing.T) {
		areUnder(t, true, []string{root + "/b"}, []string{"/nowhere", root})
	})

	t.Run("one path outside spoils them all", func(t *testing.T) {
		areUnder(t, false, []string{root, "/elsewhere"}, []string{root})
	})

	t.Run("no paths are vacuously under anything", func(t *testing.T) {
		areUnder(t, true, nil, []string{root})
	})

	t.Run("no roots contain nothing", func(t *testing.T) {
		areUnder(t, false, []string{root}, nil)
	})

	t.Run("a relative step is resolved", func(t *testing.T) {
		areUnder(t, true, []string{root + "/nested/.."}, []string{root})
	})

	t.Run("a relative step that escapes is caught", func(t *testing.T) {
		areUnder(t, false, []string{root + "/.."}, []string{root})
	})

	t.Run("a tilde is expanded on either side", func(t *testing.T) {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skip("no home directory")
		}
		areUnder(t, true, []string{home}, []string{"~"})
		areUnder(t, true, []string{"~/somewhere"}, []string{home})
	})

	t.Run("an empty path is under nothing", func(t *testing.T) {
		areUnder(t, false, []string{""}, []string{root})
	})

	t.Run("another user's home is under nothing", func(t *testing.T) {
		here, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		areUnder(t, false, []string{"~nobody/work"}, []string{here})
	})

	t.Run("a step back out of a link lands where the link led", func(t *testing.T) {
		elsewhere := t.TempDir()
		link := filepath.Join(root, "link")
		if err := os.Symlink(elsewhere, link); err != nil {
			t.Skip(err)
		}
		areUnder(t, false, []string{link + "/.."}, []string{root})
		areUnder(t, true, []string{link + "/.."}, []string{filepath.Dir(elsewhere)})
	})

	t.Run("a link out of a root leads out of it", func(t *testing.T) {
		elsewhere := t.TempDir()
		link := filepath.Join(root, "link")
		if err := os.Symlink(elsewhere, link); err != nil {
			t.Skip(err)
		}
		areUnder(t, false, []string{link}, []string{root})
		areUnder(t, false, []string{filepath.Join(link, "nested")}, []string{root})
		areUnder(t, true, []string{link}, []string{elsewhere})
	})
}

func areUnder(t *testing.T, wanted bool, paths, roots []string) {
	t.Helper()
	if all := directories.AreAllUnder(paths, roots); all != wanted {
		t.Errorf("AreAllUnder(%q, %q) = %v, wanted %v", paths, roots, all, wanted)
	}
}

func TestWithin(t *testing.T) {
	t.Run("a relative path is taken from where it is", func(t *testing.T) {
		resolvesWithin(t, "/a", "b", "/a/b")
		resolvesWithin(t, "/a", "../b", "/a/../b")
	})

	t.Run("from here it is itself", func(t *testing.T) {
		resolvesWithin(t, ".", "b", "b")
	})

	t.Run("a path from the root or a home stands alone", func(t *testing.T) {
		resolvesWithin(t, "/a", "/b", "/b")
		resolvesWithin(t, "/a", "~/b", "~/b")
	})

	t.Run("anything taken from somewhere unknown is unknown", func(t *testing.T) {
		resolvesWithin(t, "", "b", "")
		resolvesWithin(t, "/a", "", "")
	})
}

func resolvesWithin(t *testing.T, here, path, wanted string) {
	t.Helper()
	if got := directories.Within(here, path); got != filepath.FromSlash(wanted) {
		t.Errorf("Within(%q, %q) = %q, wanted %q", here, path, got, wanted)
	}
}
