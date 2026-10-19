// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadSpecs(t *testing.T) {
	passing := "/ % ls\nallow: Every command is allowed: ls\n"

	t.Run("a file is read whatever its name", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "notes.txt")
		if err := os.WriteFile(path, []byte(passing), 0o600); err != nil {
			t.Fatal(err)
		}
		specs, err := ReadSpecs(path)
		if err != nil || len(specs) != 1 || len(specs[0].Cases) != 1 || specs[0].Cases[0].Path != path {
			t.Errorf("ReadSpecs(%q) = %+v, %v", path, specs, err)
		}
	})

	t.Run("a folder is searched for .test files", func(t *testing.T) {
		root := t.TempDir()
		path := filepath.Join(root, "basics.test")
		if err := os.WriteFile(path, []byte(passing), 0o600); err != nil {
			t.Fatal(err)
		}
		specs, err := ReadSpecs(root)
		if err != nil || len(specs) != 1 || specs[0].Cases[0].Path != path {
			t.Errorf("ReadSpecs(%q) = %+v, %v", root, specs, err)
		}
	})

	t.Run("a folder with no .test files is refused", func(t *testing.T) {
		root := t.TempDir()
		if _, err := ReadSpecs(root); err == nil || err.Error() != root+": no .test files" {
			t.Errorf("ReadSpecs(%q) error = %v", root, err)
		}
	})

	t.Run("every bad path is reported", func(t *testing.T) {
		first, second := t.TempDir(), t.TempDir()
		_, err := ReadSpecs(first, second)
		if err == nil || !strings.Contains(err.Error(), first) || !strings.Contains(err.Error(), second) {
			t.Errorf("ReadSpecs error = %v", err)
		}
	})
}
